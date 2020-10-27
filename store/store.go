package store

import (
	"fmt"
	"github.com/autom8ter/graphik/fsm"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	opts *Opts
	raft *raft.Raft
}

func New(opts ...Opt) (*Store, error) {
	options := &Opts{}
	for _, o := range opts {
		o(options)
	}
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(options.localID)
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", options.bindAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(options.bindAddr, addr, 3, 10*time.Second, os.Stderr)
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
	rft, err := raft.NewRaft(config, fsm.New(), logStore, stableStore, snapshots, transport)
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
	return &Store{
		raft: rft,
		opts: options,
	}, nil
}
