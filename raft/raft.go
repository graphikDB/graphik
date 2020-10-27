package raft

import (
	"encoding/json"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

func NewRaft(localID string, raftDir, bindAddr string, leader bool) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)
	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore := boltDB
	stableStore := boltDB
	rft, err := raft.NewRaft(config, nil, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
	}
	if leader {
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
	return rft, nil
}

type fsmStore struct{}

func (f fsmStore) Apply(log *raft.Log) interface{} {
	var c command.Command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
	case command.SET_NODE:
		input := c.Val.(*model.NodeInput)
		if input.ID != nil {
			node := dagger.NewNode(input.Type, *input.ID, input.Attributes)
			return &model.Node{
				ID:         node.ID(),
				Type:       node.Type(),
				Attributes: node.Raw(),
				Edges:      nil,
			}
		} else {
			node := dagger.NewNode(input.Type, "", input.Attributes)
			return &model.Node{
				ID:         node.ID(),
				Type:       node.Type(),
				Attributes: node.Raw(),
				Edges:      nil,
			}
		}
	case command.DELETE_NODE:
		input := c.Val.(*model.ForeignKey)
		dagger.DelNode(dagger.ForeignKey(input.Type, input.ID))
		return nil
	case command.SET_EDGE:
		input := c.Val.(*model.EdgeInput)
		from, ok := dagger.GetNode(dagger.ForeignKey(input.From.Type, input.From.ID))
		if !ok {
			return fmt.Errorf("%s.%s does not exist", input.From.Type, input.From.ID)
		}
		to, ok := dagger.GetNode(dagger.ForeignKey(input.To.Type, input.To.ID))
		if !ok {
			return fmt.Errorf("%s.%s does not exist", input.To.Type, input.To.ID)
		}
		edge, err := dagger.NewEdge(input.Node.Type, nullString(input.Node.ID), input.Node.Attributes, from, to)
		if err != nil {
			return err
		}
		return &model.Edge{
			Node: &model.Node{
				ID:         edge.ID(),
				Type:       edge.Type(),
				Attributes: edge.Node().Raw(),
				Edges:      nil,
			},
			From: &model.Node{
				ID:         from.ID(),
				Type:       from.Type(),
				Attributes: from.Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				ID:         to.ID(),
				Type:       to.Type(),
				Attributes: to.Raw(),
				Edges:      nil,
			},
		}
	}
	return nil
}

func (f *fsmStore) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *fsmStore) Restore(closer io.ReadCloser) error {
	return dagger.ImportJSON(closer)
}

func (f *fsmStore) Persist(sink raft.SnapshotSink) error {
	if err := dagger.ExportJSON(sink); err != nil {
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *fsmStore) Release() {}

func nullString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
