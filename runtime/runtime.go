package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/config"
	"github.com/autom8ter/graphik/express"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/storage"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	raftTimeout         = 10 * time.Second
	ChangeStreamChannel = "changes"
)

type Runtime struct {
	machine *machine.Machine
	config  *config.Config
	mu      sync.RWMutex
	raft    *raft.Raft
	graph   *graph.Graph
	close   sync.Once
}

func New(ctx context.Context, cfg *apipb.Config) (*Runtime, error) {
	os.MkdirAll(cfg.GetRaft().GetStoragePath(), 0700)
	m := machine.New(ctx)
	rconfig := raft.DefaultConfig()
	rconfig.LocalID = raft.ServerID(cfg.GetRaft().GetNodeId())
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", cfg.GetRaft().GetBind())
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve raft bind address")
	}
	transport, err := raft.NewTCPTransport(cfg.GetRaft().GetBind(), addr, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raft transport")
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(cfg.GetRaft().GetStoragePath(), 2, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create snapshot store")
	}
	logStore, err := storage.NewLogStore(filepath.Join(cfg.GetRaft().GetStoragePath(), "logs.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bolt store")
	}
	snapshotStore, err := storage.NewSnapshotStore(filepath.Join(cfg.GetRaft().GetStoragePath(), "snapshots.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bolt store")
	}
	c, err := config.New(cfg.GetRuntime())
	if err != nil {
		return nil, err
	}
	g, err := graph.New(filepath.Join(cfg.GetRaft().GetStoragePath(), "graph.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create graph")
	}
	s := &Runtime{
		machine: m,
		config:  c,
		raft:    nil,
		mu:      sync.RWMutex{},
		graph:   g,
		close:   sync.Once{},
	}
	rft, err := raft.NewRaft(rconfig, s, logStore, snapshotStore, snapshots, transport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raft")
	}
	s.raft = rft
	if cfg.GetRaft().GetJoin() == "" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      rconfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		conn, err := grpc.DialContext(ctx, cfg.Raft.Join, grpc.WithInsecure())
		if err != nil {
			return nil, errors.Wrap(err, "failed to join cluster")
		}
		client := apipb.NewGraphServiceClient(conn)
		_, err = client.JoinCluster(ctx, &apipb.RaftNode{
			NodeId:  cfg.GetRaft().GetNodeId(),
			Address: cfg.GetRaft().GetBind(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to join cluster")
		}
	}

	s.machine.Go(func(routine machine.Routine) {
		logger.Info("refreshing jwks")
		if err := s.config.RefreshKeys(); err != nil {
			logger.Error("failed to refresh keys", zap.Error(errors.WithStack(err)))
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
	return s, nil
}

func (s *Runtime) execute(cmd *apipb.StateChange) (*apipb.Log, error) {
	if state := s.raft.State(); state != raft.Leader {
		return nil, fmt.Errorf("not leader: %s", state.String())
	}
	now := time.Now()
	//defer func() {
	//	logger.Info("executing raft mutation", zap.Int64("duration(ms)", time.Since(now).Milliseconds()))
	//}()
	bits, err := proto.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	future := s.raft.Apply(bits, raftTimeout)
	if err := future.Error(); err != nil {
		return nil, err
	}
	logger.Info("raft mutation completed", zap.Int64("duration(ms)", time.Since(now).Milliseconds()))
	if err, ok := future.Response().(error); ok {
		return nil, err
	}
	return future.Response().(*apipb.Log), nil
}

func (s *Runtime) Close() error {
	defer s.machine.Close()
	if err := s.raft.Shutdown().Error(); err != nil {
		return err
	}
	return nil
}

func (s *Runtime) Machine() *machine.Machine {
	return s.machine
}

func (s *Runtime) Config() *config.Config {
	return s.config
}

func (a *Runtime) Authorize(intercept *apipb.RequestIntercept) (bool, error) {
	return express.Eval(a.config.AuthPrograms(), intercept)
}

func (r *Runtime) Go(fn machine.Func) {
	r.machine.Go(fn)
}

func (r *Runtime) Cancel() {
	r.machine.Cancel()
}

func (r *Runtime) Wait() {
	r.machine.Wait()
}
