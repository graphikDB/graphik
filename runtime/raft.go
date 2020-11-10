package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/values"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
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
	var c apipb.Command
	if err := proto.Unmarshal(log.Data, &c); err != nil {
		return fmt.Errorf("failed to decode command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case apipb.Op_SET_AUTH:
		var auth apipb.AuthConfig
		if err := ptypes.UnmarshalAny(c.Val, &auth); err != nil {
			return err
		}
		if err := f.auth.Override(&auth); err != nil {
			return errors.Wrap(err, "failed to override auth")
		}
		return f.auth.Raw()
	default:
		return fmt.Errorf("unsupported command: %v", c.Op)
	}
}

func (f *Runtime) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *Runtime) Restore(closer io.ReadCloser) error {
	export := &apipb.Export{}
	bits, err := ioutil.ReadAll(closer)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(bits, export); err != nil {
		return err
	}
	nodes := values.ValueSet{}
	nodes.FromStructs(export.GetNodes())
	edges := values.ValueSet{}
	edges.FromStructs(export.GetEdges())
	f.mu.Lock()
	defer f.mu.Unlock()
	f.graph.SetNodes(nodes)
	f.graph.SetEdges(edges)
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	export := &apipb.Export{
		Nodes: f.graph.AllNodes().Structs(),
		Edges: f.graph.AllEdges().Structs(),
	}
	bits, err := proto.Marshal(export)
	if err != nil {
		return err
	}
	_, err = sink.Write(bits)
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
