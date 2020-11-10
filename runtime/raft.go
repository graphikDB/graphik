package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/golang/protobuf/proto"
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

func (f *Runtime) apply(log *raft.Log) (*apipb.RaftLog, error) {
	var c apipb.Command
	if err := proto.Unmarshal(log.Data, &c); err != nil {
		return nil, fmt.Errorf("failed to decode command: %s", err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch c.Op {
	case apipb.Op_SET_AUTH:
		if err := f.auth.Override(c.Val.GetAuth()); err != nil {
			return nil, errors.Wrap(err, "failed to override auth")
		}
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Auth{Auth: f.auth.Raw()},
		}, nil
	case apipb.Op_CREATE_NODE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Node{Node: f.graph.SetNode(c.Val.GetNode())},
		}, nil
	case apipb.Op_CREATE_EDGE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Edge{Edge: f.graph.SetEdge(c.Val.GetEdge())},
		}, nil
	case apipb.Op_CREATE_NODES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Nodes{Nodes: f.graph.SetNodes(c.Val.GetNodes().GetNodes())},
		}, nil
	case apipb.Op_CREATE_EDGES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Edges{Edges: f.graph.SetEdges(c.Val.GetEdges().GetEdges())},
		}, nil

	case apipb.Op_PATCH_NODE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Node{Node: f.graph.PatchNode(c.Val.GetNode())},
		}, nil
	case apipb.Op_PATCH_EDGE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Edge{Edge: f.graph.PatchEdge(c.Val.GetEdge())},
		}, nil
	case apipb.Op_PATCH_NODES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Nodes{Nodes: f.graph.PatchNodes(c.Val.GetNodes().GetNodes())},
		}, nil
	case apipb.Op_PATCH_EDGES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Edges{Edges: f.graph.PatchEdges(c.Val.GetEdges().GetEdges())},
		}, nil

	case apipb.Op_DELETE_NODE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Counter{Counter: f.graph.DeleteNode(c.Val.GetPath())},
		}, nil
	case apipb.Op_DELETE_EDGE:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Counter{Counter: f.graph.DeleteEdge(c.Val.GetPath())},
		}, nil
	case apipb.Op_DELETE_NODES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Counter{Counter: f.graph.DeleteNodes(c.Val.GetPaths().GetPaths())},
		}, nil
	case apipb.Op_DELETE_EDGES:
		return &apipb.RaftLog{
			Log: &apipb.RaftLog_Counter{Counter: f.graph.DeleteEdges(c.Val.GetPaths().GetPaths())},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported command: %v", c.Op)
	}
}

func (f *Runtime) Apply(log *raft.Log) interface{} {
	res, err := f.apply(log)
	if err != nil {
		return err
	}
	return res
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
	f.mu.Lock()
	defer f.mu.Unlock()
	f.graph.SetNodes(export.GetNodes().GetNodes())
	f.graph.SetEdges(export.GetEdges().GetEdges())
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	export := &apipb.Export{
		Nodes: f.graph.AllNodes(),
		Edges: f.graph.AllEdges(),
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
