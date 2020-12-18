package raft

import (
	"fmt"
	"github.com/graphikDB/graphik/logger"
	"github.com/graphikDB/graphik/raft/fsm"
	"github.com/graphikDB/graphik/raft/storage"
	"github.com/hashicorp/raft"
	"net"
	"os"
	"path/filepath"
)

type Raft struct {
	raft *raft.Raft
	opts *Options
}

func NewRaft(fsm *fsm.FSM, opts ...Opt) (*Raft, error) {
	if err := fsm.Validate(); err != nil {
		return nil, err
	}
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	options.setDefaults()
	fmt.Println(options.raftDir)
	fmt.Println(options.peerID)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(options.peerID)
	addr, err := net.ResolveTCPAddr("tcp", options.listenAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(options.listenAddr, addr, options.maxPool, options.timeout, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(options.raftDir, options.retainSnapshots, os.Stderr)
	if err != nil {
		return nil, err
	}
	strg, err := storage.NewStorage(filepath.Join(options.raftDir, "raft.db"))
	if err != nil {
		return nil, err
	}
	ra, err := raft.NewRaft(config, fsm, strg, strg, snapshots, transport)
	if err != nil {
		return nil, err
	}
	if options.isLeader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		logger.Info("bootstrapping raft cluster as leader")
		ra.BootstrapCluster(configuration)
	}
	return &Raft{
		opts: options,
		raft: ra,
	}, nil
}

func (r *Raft) State() raft.RaftState {
	return r.raft.State()
}

func (s *Raft) Servers() ([]raft.Server, error) {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return nil, err
	}
	return configFuture.Configuration().Servers, nil
}

func (s *Raft) Join(nodeID, addr string) error {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				// already a member
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
	return nil
}

func (s *Raft) LeaderAddr() string {
	return string(s.raft.Leader())
}

func (s *Raft) Stats() map[string]string {
	return s.raft.Stats()
}

func (s *Raft) Apply(bits []byte) (interface{}, error) {
	f := s.raft.Apply(bits, s.opts.timeout)
	if err := f.Error(); err != nil {
		return nil, err
	}
	resp := f.Response()
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp, nil
}

func (r *Raft) Close() error {
	return r.raft.Shutdown().Error()
}
