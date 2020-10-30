package store

import (
	"fmt"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/generic"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
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

type Store struct {
	opts  *Opts
	raft  *raft.Raft
	mu    sync.RWMutex
	nodes *generic.Nodes
	edges *generic.Edges
	close sync.Once
}

func New(opts ...Opt) (*Store, error) {
	options := &Opts{}
	for _, o := range opts {
		o(options)
	}
	if options.localID == "" {
		id, _ := getID()
		options.localID = id
	}
	if options.raftDir == "" {
		os.MkdirAll(defaultDir, 0700)
		options.raftDir = defaultDir
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
	s := &Store{
		opts:  options,
		mu:    sync.RWMutex{},
		nodes: generic.NewNodes(),
		edges: generic.NewEdges(),
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
	return s, nil
}

func (s *Store) Execute(cmd *command.Command) (interface{}, error) {
	if state := s.raft.State(); state != raft.Leader {
		return nil, fmt.Errorf("not leader: %s", state.String())
	}
	lg, err := cmd.Log()
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

func getID() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var ip net.IP
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
		}
	}
	return ip.String(), nil
}

func (s *Store) Close() error {
	return s.raft.Shutdown().Error()
}
