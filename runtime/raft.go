package runtime

import (
	"bytes"
	"encoding/gob"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/logger"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
)

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

func (f *Runtime) Apply(log *raft.Log) interface{} {
	var c graph.Command
	buf := bytes.NewBuffer(log.Data)
	if err := gob.NewDecoder(buf).Decode(&c); err != nil {
		return fmt.Errorf("failed to decode command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case graph.Op_SET_AUTH:
		if err := f.auth.Override(c.Val.(*apipb.AuthConfig)); err != nil {
			return errors.Wrap(err, "failed to override auth")
		}
		return f.auth.Raw()
	case graph.Op_CREATE_NODES:
		values := c.Val.(*graph.ValueSet)
		f.graph.Nodes().SetAll(*values)
		return *values
	case graph.Op_PATCH_NODES:
		var nodes = graph.ValueSet{}
		for _, val := range *c.Val.(*graph.ValueSet) {
			if !f.graph.Nodes().Exists(val.PathString()) {
				return errors.Errorf("node %s does not exist", val.PathString())
			}
			n := f.graph.Nodes().Patch(c.Timestamp, val)
			nodes = append(nodes, n)
		}
		return nodes
	case graph.Op_DELETE_NODES:
		deleted := 0
		for _, val := range c.Val.([]string) {
			if f.graph.Nodes().Delete(val) {
				deleted += 1
			}
		}
		return deleted
	case graph.Op_CREATE_EDGES:
		values := c.Val.(*graph.ValueSet)
		f.graph.Edges().SetAll(*values)
		return *values
	case graph.Op_PATCH_EDGES:
		var edges = graph.ValueSet{}
		for _, val := range *c.Val.(*graph.ValueSet) {
			if !f.graph.Edges().Exists(val.PathString()) {
				return errors.Errorf("edge %s does not exist", val.PathString())
			}
			edges = append(edges, f.graph.Edges().Patch(c.Timestamp, val))
		}
		return edges

	case graph.Op_DELETE_EDGES:
		deleted := 0
		for _, val := range c.Val.([]string) {
			if f.graph.Edges().Exists(val) {
				f.graph.Edges().Delete(val)
				deleted += 1
			}
		}
		return deleted
	default:
		return fmt.Errorf("unsupported command: %v", c.Op)
	}
}

func (f *Runtime) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *Runtime) Restore(closer io.ReadCloser) error {
	export := &graph.Export{}
	if err := gob.NewDecoder(closer).Decode(export); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.graph.Nodes().SetAll(export.Nodes)
	f.graph.Edges().SetAll(export.Edges)
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	export := &graph.Export{
		Nodes: f.graph.Nodes().All(),
		Edges: f.graph.Edges().All(),
	}
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(export); err != nil {
		return err
	}
	_, err := sink.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if err := sink.Close(); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (f *Runtime) Release() {}
