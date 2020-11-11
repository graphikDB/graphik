package service

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/express"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/runtime"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"time"
)

type Graph struct {
	runtime *runtime.Runtime
}

func NewGraph(runtime *runtime.Runtime) *Graph {
	return &Graph{runtime: runtime}
}

func (g *Graph) Me(ctx context.Context, empty *empty.Empty) (*apipb.Node, error) {
	n := g.runtime.NodeContext(ctx)
	return n, nil
}

func (g *Graph) CreateNode(ctx context.Context, node *apipb.Node) (*apipb.Node, error) {
	return g.runtime.CreateNode(node)
}

func (g *Graph) CreateNodes(ctx context.Context, nodes *apipb.Nodes) (*apipb.Nodes, error) {
	return g.runtime.CreateNodes(nodes)
}

func (g *Graph) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	return g.runtime.Node(path)
}

func (g *Graph) SearchNodes(ctx context.Context, filter *apipb.TypeFilter) (*apipb.Nodes, error) {
	return g.runtime.Nodes(filter)
}

func (g *Graph) PatchNode(ctx context.Context, node *apipb.Node) (*apipb.Node, error) {
	return g.runtime.PatchNode(node)
}

func (g *Graph) PatchNodes(ctx context.Context, nodes *apipb.Nodes) (*apipb.Nodes, error) {
	return g.runtime.PatchNodes(nodes)
}

func (g *Graph) DelNode(ctx context.Context, path *apipb.Path) (*apipb.Counter, error) {
	return g.runtime.DelNode(path)
}

func (g *Graph) DelNodes(ctx context.Context, paths *apipb.Paths) (*apipb.Counter, error) {
	return g.runtime.DelNodes(paths)
}

func (g *Graph) CreateEdge(ctx context.Context, edge *apipb.Edge) (*apipb.Edge, error) {
	return g.runtime.CreateEdge(edge)
}

func (g *Graph) CreateEdges(ctx context.Context, edges *apipb.Edges) (*apipb.Edges, error) {
	return g.runtime.CreateEdges(edges)
}

func (g *Graph) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	return g.runtime.Edge(path)
}

func (g *Graph) SearchEdges(ctx context.Context, filter *apipb.TypeFilter) (*apipb.Edges, error) {
	return g.runtime.Edges(filter), nil
}

func (g *Graph) PatchEdge(ctx context.Context, edge *apipb.Edge) (*apipb.Edge, error) {
	return g.runtime.PatchEdge(edge)
}

func (g *Graph) PatchEdges(ctx context.Context, edges *apipb.Edges) (*apipb.Edges, error) {
	return g.runtime.PatchEdges(edges)
}

func (g *Graph) DelEdge(ctx context.Context, path *apipb.Path) (*apipb.Counter, error) {
	return g.runtime.DelEdge(path)
}

func (g *Graph) DelEdges(ctx context.Context, paths *apipb.Paths) (*apipb.Counter, error) {
	return g.runtime.DelEdges(paths)
}

func (g *Graph) ChangeStream(filter *apipb.ChangeFilter, server apipb.GraphService_ChangeStreamServer) error {
	filterFunc := func(msg interface{}) bool {
		result, err := express.Eval(filter.Expressions, msg)
		if err != nil {
			logger.Error("subscription filter failure", zap.Error(err))
			return false
		}
		return result
	}
	if err := g.runtime.Machine().PubSub().SubscribeFilter(server.Context(), runtime.ChangeStreamChannel, filterFunc, func(msg interface{}) {
		if err, ok := msg.(error); ok && err != nil {
			logger.Error("failed to send subscription", zap.Error(err))
			return
		}
		if val, ok := msg.(*apipb.StateChange); ok {
			if err := server.Send(val); err != nil {
				logger.Error("failed to send subscription", zap.Error(err))
				return
			}
		}
	}); err != nil {
		return err
	}
	return nil
}

func (g *Graph) Publish(ctx context.Context, message *apipb.OutboundMessage) (*empty.Empty, error) {
	return nil, g.runtime.Machine().PubSub().Publish(message.Channel, &apipb.Message{
		Channel:   message.Channel,
		Data:      message.Data,
		Sender:    g.runtime.NodeContext(ctx).Path,
		Timestamp: time.Now().UnixNano(),
	})
}

func (g *Graph) Subscribe(filter *apipb.ChannelFilter, server apipb.GraphService_SubscribeServer) error {
	filterFunc := func(msg interface{}) bool {
		result, err := express.Eval(filter.Expressions, msg)
		if err != nil {
			logger.Error("subscription filter failure", zap.Error(err))
			return false
		}
		return result
	}
	if err := g.runtime.Machine().PubSub().SubscribeFilter(server.Context(), filter.Channel, filterFunc, func(msg interface{}) {
		if err, ok := msg.(error); ok && err != nil {
			logger.Error("failed to send subscription", zap.Error(err))
			return
		}
		if val, ok := msg.(*apipb.Message); ok {
			if err := server.Send(val); err != nil {
				logger.Error("failed to send subscription", zap.Error(err))
				return
			}
		}
	}); err != nil {
		return err
	}
	return nil
}

func (g *Graph) EdgesFrom(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	return g.runtime.EdgesFrom(filter)
}

func (g *Graph) EdgesTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	return g.runtime.EdgesTo(filter)
}
