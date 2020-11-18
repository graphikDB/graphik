package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"time"
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

func (f *Runtime) apply(ctx context.Context, log *raft.Log) (*apipb.StateChange, error) {
	var c = &apipb.StateChange{}
	if err := proto.Unmarshal(log.Data, c); err != nil {
		return nil, fmt.Errorf("failed to decode command: %s", err.Error())
	}
	switch c.Op {
	case apipb.Op_CREATE_NODE:
		n := c.Mutation.GetNodeConstructor()
		node, err := f.graph.SetNode(ctx, &apipb.Node{
			Path:       n.Path,
			Attributes: n.Attributes,
			CreatedAt:  c.Timestamp,
			UpdatedAt:  c.Timestamp,
		})
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Node{Node: node},
		}
	case apipb.Op_CREATE_EDGE:
		e := c.Mutation.GetEdgeConstructor()
		edge, err := f.graph.SetEdge(ctx, &apipb.Edge{
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
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Edge{Edge: edge},
		}
	case apipb.Op_CREATE_NODES:
		n := c.Mutation.GetNodeConstructors()
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
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Nodes{Nodes: res},
		}
	case apipb.Op_CREATE_EDGES:
		e := c.Mutation.GetEdgeConstructors()
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
		res, err := f.graph.SetEdges(ctx, edges)
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Edges{Edges: res},
		}
	case apipb.Op_PATCH_NODE:
		res, err := f.graph.PatchNode(ctx, c.Mutation.GetPatch())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Node{Node: res},
		}
	case apipb.Op_PATCH_EDGE:
		res, err := f.graph.PatchEdge(ctx, c.Mutation.GetPatch())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Edge{Edge: res},
		}
	case apipb.Op_PATCH_NODES:
		res, err := f.graph.PatchNodes(ctx, c.Mutation.GetPatches())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Nodes{Nodes: res},
		}
	case apipb.Op_PATCH_EDGES:
		res, err := f.graph.PatchEdges(ctx, c.Mutation.GetPatches())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Edges{Edges: res},
		}
	case apipb.Op_DELETE_NODE:
		res, err := f.graph.DeleteNode(ctx, c.Mutation.GetPath())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Empty{Empty: res},
		}
	case apipb.Op_DELETE_EDGE:
		res, err := f.graph.DeleteEdge(ctx, c.Mutation.GetPath())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Empty{Empty: res},
		}
	case apipb.Op_DELETE_NODES:
		res, err := f.graph.DeleteNodes(ctx, c.Mutation.GetPaths().GetPaths())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Empty{Empty: res},
		}
	case apipb.Op_DELETE_EDGES:
		res, err := f.graph.DeleteEdges(ctx, c.Mutation.GetPaths().GetPaths())
		if err != nil {
			return nil, err
		}
		c.Mutation = &apipb.Mutation{
			Object: &apipb.Mutation_Empty{Empty: res},
		}
	default:
		return nil, fmt.Errorf("unsupported command: %v", c.Op)
	}
	if err := f.machine.PubSub().Publish(ChangeStreamChannel, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (f *Runtime) Apply(log *raft.Log) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := f.apply(ctx, log)
	if err != nil {
		return err
	}
	return res
}

func (f *Runtime) Snapshot() (raft.FSMSnapshot, error) {
	return f, nil
}

func (f *Runtime) Restore(closer io.ReadCloser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	export := &apipb.Graph{}
	bits, err := ioutil.ReadAll(closer)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(bits, export); err != nil {
		return err
	}
	_, err = f.Import(ctx, export)
	if err != nil {
		return err
	}
	return nil
}

func (f *Runtime) Persist(sink raft.SnapshotSink) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	export, err := f.Export(ctx)
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
