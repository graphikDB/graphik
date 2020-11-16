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

func (f *Runtime) apply(log *raft.Log) (*apipb.Log, error) {
	var c = &apipb.StateChange{}
	if err := proto.Unmarshal(log.Data, c); err != nil {
		return nil, fmt.Errorf("failed to decode command: %s", err.Error())
	}
	switch c.Op {
	case apipb.Op_SET_AUTH:
		if err := f.auth.Override(c.Log.GetAuth()); err != nil {
			return nil, errors.Wrap(err, "failed to override auth")
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Auth{Auth: f.auth.Raw()},
		}
	case apipb.Op_CREATE_NODE:
		n := c.Log.GetNodeConstructor()
		node, err := f.graph.SetNode(&apipb.Node{
			Path:       n.Path,
			Attributes: n.Attributes,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Node{Node: node},
		}
	case apipb.Op_CREATE_EDGE:
		e := c.Log.GetEdgeConstructor()
		edge, err := f.graph.SetEdge(&apipb.Edge{
			Path:       e.Path,
			Attributes: e.Attributes,
			Cascade:    e.Cascade,
			From:       e.From,
			To:         e.To,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Edge{Edge: edge},
		}
	case apipb.Op_CREATE_NODES:
		n := c.Log.GetNodeConstructors()
		var nodes []*apipb.Node
		for _, node := range n.GetNodes() {
			nodes = append(nodes, &apipb.Node{
				Path:       node.Path,
				Attributes: node.Attributes,
				CreatedAt:  c.Timestamp,
				UpdatedAt:  c.Timestamp,
			})
		}
		res, err := f.graph.SetNodes(nodes)
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Nodes{Nodes: res},
		}
	case apipb.Op_CREATE_EDGES:
		e := c.Log.GetEdgeConstructors()
		var edges []*apipb.Edge
		for _, edge := range e.GetEdges() {
			edges = append(edges, &apipb.Edge{
				Path:       edge.Path,
				Attributes: edge.Attributes,
				Cascade:    edge.Cascade,
				From:       edge.From,
				To:         edge.To,
				CreatedAt:  c.Timestamp,
				UpdatedAt:  c.Timestamp,
			})
		}
		res, err := f.graph.SetEdges(edges)
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Edges{Edges: res},
		}
	case apipb.Op_PATCH_NODE:
		res, err := f.graph.PatchNode(c.Log.GetPatch())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Node{Node: res},
		}
	case apipb.Op_PATCH_EDGE:
		res, err := f.graph.PatchEdge(c.Log.GetPatch())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Edge{Edge: res},
		}
	case apipb.Op_PATCH_NODES:
		res, err := f.graph.PatchNodes(c.Log.GetPatches())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Nodes{Nodes: res},
		}
	case apipb.Op_PATCH_EDGES:
		res, err := f.graph.PatchEdges(c.Log.GetPatches())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Edges{Edges: res},
		}
	case apipb.Op_DELETE_NODE:
		res, err := f.graph.DeleteNode(c.Log.GetPath())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Counter{Counter: res},
		}
	case apipb.Op_DELETE_EDGE:
		res, err := f.graph.DeleteEdge(c.Log.GetPath())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Counter{Counter: res},
		}
	case apipb.Op_DELETE_NODES:
		res, err := f.graph.DeleteNodes(c.Log.GetPaths().GetPaths())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Counter{Counter: res},
		}
	case apipb.Op_DELETE_EDGES:
		res, err := f.graph.DeleteEdges(c.Log.GetPaths().GetPaths())
		if err != nil {
			return nil, err
		}
		c.Log = &apipb.Log{
			Log: &apipb.Log_Counter{Counter: res},
		}
	default:
		return nil, fmt.Errorf("unsupported command: %v", c.Op)
	}
	if err := f.machine.PubSub().Publish(ChangeStreamChannel, c); err != nil {
		return nil, err
	}
	return c.Log, nil
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
	export := &apipb.Graph{}
	bits, err := ioutil.ReadAll(closer)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(bits, export); err != nil {
		return err
	}
	_, err = f.Import(export)
	if err != nil {
		return err
	}
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	export, err := f.Export()
	if err != nil {
		return err
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
