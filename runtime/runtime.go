package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/auth"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
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
	raftTimeout = 10 * time.Second
)

type Runtime struct {
	machine *machine.Machine
	auth    *auth.Auth
	raft    *raft.Raft
	mu      sync.RWMutex
	nodes   *graph.Nodes
	edges   *graph.Edges
	close   sync.Once
}

func New(ctx context.Context, cfg *apipb.Config) (*Runtime, error) {
	os.MkdirAll(cfg.GetRaft().GetStoragePath(), 0700)
	m := machine.New(ctx)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(cfg.GetRaft().GetNodeId())
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
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(cfg.GetRaft().GetStoragePath(), "raft.db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bolt store")
	}
	logStore := boltDB
	stableStore := boltDB
	edges := graph.NewEdges()
	nodes := graph.NewNodes(edges)
	a, err := auth.New(cfg.Auth)
	if err != nil {
		return nil, err
	}
	s := &Runtime{
		machine: m,
		auth:    a,
		raft:    nil,
		mu:      sync.RWMutex{},
		nodes:   nodes,
		edges:   edges,
		close:   sync.Once{},
	}
	rft, err := raft.NewRaft(config, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create raft")
	}
	s.raft = rft
	if cfg.GetRaft().GetJoin() == "" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
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
		client := apipb.NewPrivateServiceClient(conn)
		_, err = client.JoinCluster(ctx, &apipb.JoinClusterRequest{
			RaftNode: &apipb.RaftNode{
				NodeId:  cfg.GetRaft().GetNodeId(),
				Address: cfg.GetRaft().GetBind(),
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to join cluster")
		}
	}

	s.machine.Go(func(routine machine.Routine) {
		logger.Info("refreshing jwks")
		if err := s.auth.RefreshKeys(); err != nil {
			logger.Error("failed to refresh keys", zap.Error(errors.WithStack(err)))
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
	return s, nil
}

func (s *Runtime) execute(cmd *apipb.Command) (interface{}, error) {
	if state := s.raft.State(); state != raft.Leader {
		return nil, fmt.Errorf("not leader: %s", state.String())
	}
	future := s.raft.ApplyLog(*cmd.Log(), raftTimeout)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return future.Response(), nil
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

func (s *Runtime) Auth() *auth.Auth {
	return s.auth
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
