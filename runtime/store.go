package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/graphik/api/model"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/zap"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	raftTimeout = 10 * time.Second
	defaultDir  = "/tmp/graphik"
)

type Runtime struct {
	opts  *Opts
	raft  *raft.Raft
	mu    sync.RWMutex
	nodes *generic.Nodes
	edges *generic.Edges
	close sync.Once
}

func New(opts ...Opt) (*Runtime, error) {
	options := &Opts{}
	for _, o := range opts {
		o(options)
	}
	if options.jwks == nil || len(options.jwks.List()) == 0 {
		return nil, fmt.Errorf("empty jwks")
	}
	if options.localID == "" {
		options.localID = "default"
	}
	if options.raftDir == "" {
		os.MkdirAll(defaultDir, 0700)
		options.raftDir = defaultDir
	}
	if options.machine == nil {
		options.machine = machine.New(context.Background())
	}
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(options.localID)
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", options.bindAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(options.bindAddr, addr, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(options.raftDir, 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(options.raftDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore := boltDB
	stableStore := boltDB
	edges := generic.NewEdges()
	nodes := generic.NewNodes(edges)
	s := &Runtime{
		opts:  options,
		mu:    sync.RWMutex{},
		nodes: nodes,
		edges: edges,
		close: sync.Once{},
	}
	rft, err := raft.NewRaft(config, s, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
	}
	if options.leader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		rft.BootstrapCluster(configuration)
	}
	s.raft = rft
	s.opts.machine.Go(func(routine machine.Routine) {
		logger.Info("refreshing jwks")
		if err := s.opts.jwks.RefreshKeys(); err != nil {
			logger.Error("failed to refresh keys", zap.Error(err))
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
	return s, nil
}

func (s *Runtime) Execute(cmd *model.Command) (interface{}, error) {
	if state := s.raft.State(); state != raft.Leader {
		return nil, fmt.Errorf("not leader: %s", state.String())
	}
	lg, err := logCmd(cmd)
	if err != nil {
		return nil, err
	}
	future := s.raft.ApplyLog(lg, raftTimeout)
	err = future.Error()
	if err := future.Error(); err != nil {
		return nil, err
	}
	return future.Response(), nil
}

func (s *Runtime) Close() error {
	return s.raft.Shutdown().Error()
}

func (s *Runtime) Machine() *machine.Machine {
	return s.opts.machine
}

func logCmd(cmd *model.Command) (raft.Log, error) {
	cmd.Timestamp = time.Now()
	bits, err := json.Marshal(cmd)
	if err != nil {
		return raft.Log{}, err
	}
	return raft.Log{
		Data: bits,
	}, nil
}

func (s *Runtime) join(nodeID, addr string) error {
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
