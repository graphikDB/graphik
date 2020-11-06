package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/graphik/jwks"
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
	jwks    *jwks.Auth
	raft    *raft.Raft
	mu      sync.RWMutex
	nodes   *generic.Nodes
	edges   *generic.Edges
	close   sync.Once
}

func New(ctx context.Context, cfg *apipb.Config) (*Runtime, error) {
	if cfg.Auth == nil || len(cfg.Auth.JwksSources) == 0 {
		return nil, fmt.Errorf("empty jwks")
	}
	sources, err := jwks.New(cfg.GetAuth().GetJwksSources())
	if err != nil {
		return nil, err
	}
	os.MkdirAll(cfg.GetRaft().GetStoragePath(), 0700)
	m := machine.New(ctx)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(cfg.GetRaft().GetNodeId())
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", cfg.GetRaft().GetBind())
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(cfg.GetRaft().GetBind(), addr, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(cfg.GetRaft().GetStoragePath(), 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(cfg.GetRaft().GetStoragePath(), "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore := boltDB
	stableStore := boltDB
	edges := generic.NewEdges()
	nodes := generic.NewNodes(edges)

	s := &Runtime{
		machine: m,
		jwks:    sources,
		raft:    nil,
		mu:      sync.RWMutex{},
		nodes:   nodes,
		edges:   edges,
		close:   sync.Once{},
	}
	rft, err := raft.NewRaft(config, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
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
			NodeId:  cfg.GetRaft().GetNodeId(),
			Address: cfg.GetRaft().GetBind(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to join cluster")
		}
	}

	s.machine.Go(func(routine machine.Routine) {
		logger.Info("refreshing jwks")
		if err := s.jwks.RefreshKeys(); err != nil {
			logger.Error("failed to refresh keys", zap.Error(err))
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
	return s.raft.Shutdown().Error()
}

func (s *Runtime) Machine() *machine.Machine {
	return s.machine
}

func (s *Runtime) JWKS() *jwks.Auth {
	return s.jwks
}

func (s *Runtime) JoinNode(nodeID, addr string) error {
	logger.Info("received join request for remote node",
		zap.String("node", nodeID),
		zap.String("address", addr),
	)
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				logger.Info("node already member of cluster, ignoring join request",
					zap.String("node", nodeID),
					zap.String("address", addr),
				)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}
	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	logger.Info("node joined successfully",
		zap.String("node", nodeID),
		zap.String("address", addr),
	)
	return nil
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
