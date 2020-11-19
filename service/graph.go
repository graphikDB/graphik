package service

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/runtime"
	"github.com/autom8ter/graphik/vm"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Graph struct {
	runtime *runtime.Runtime
}

func NewGraph(runtime *runtime.Runtime) *Graph {
	return &Graph{runtime: runtime}
}

func (s *Graph) JoinCluster(ctx context.Context, request *apipb.RaftNode) (*empty.Empty, error) {
	if err := s.runtime.JoinNode(request.GetNodeId(), request.GetAddress()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *Graph) Ping(ctx context.Context, r *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (g *Graph) Me(ctx context.Context, filter *apipb.MeFilter) (*apipb.NodeDetail, error) {
	n := g.runtime.NodeContext(ctx)
	return g.runtime.GetNodeDetail(ctx, &apipb.NodeDetailFilter{
		Path:      n.Path,
		EdgesFrom: filter.EdgesFrom,
		EdgesTo:   filter.EdgesTo,
	})
}

func (g *Graph) CreateNode(ctx context.Context, node *apipb.NodeConstructor) (*apipb.Node, error) {
	return g.runtime.CreateNode(ctx, node)
}

func (g *Graph) CreateNodes(ctx context.Context, nodes *apipb.NodeConstructors) (*apipb.Nodes, error) {
	return g.runtime.CreateNodes(ctx, nodes)
}

func (g *Graph) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	return g.runtime.Node(ctx, path)
}

func (g *Graph) SearchNodes(ctx context.Context, filter *apipb.Filter) (*apipb.Nodes, error) {
	return g.runtime.Nodes(ctx, filter)
}

func (g *Graph) PatchNode(ctx context.Context, node *apipb.Patch) (*apipb.Node, error) {
	return g.runtime.PatchNode(ctx, node)
}

func (g *Graph) PatchNodes(ctx context.Context, nodes *apipb.Patches) (*apipb.Nodes, error) {
	return g.runtime.PatchNodes(ctx, nodes)
}

func (g *Graph) DelNode(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	return g.runtime.DelNode(ctx, path)
}

func (g *Graph) DelNodes(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	return g.runtime.DelNodes(ctx, paths)
}

func (g *Graph) CreateEdge(ctx context.Context, edge *apipb.EdgeConstructor) (*apipb.Edge, error) {
	return g.runtime.CreateEdge(ctx, edge)
}

func (g *Graph) CreateEdges(ctx context.Context, edges *apipb.EdgeConstructors) (*apipb.Edges, error) {
	return g.runtime.CreateEdges(ctx, edges)
}

func (g *Graph) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	return g.runtime.Edge(ctx, path)
}

func (g *Graph) SearchEdges(ctx context.Context, filter *apipb.Filter) (*apipb.Edges, error) {
	return g.runtime.Edges(ctx, filter)
}

func (g *Graph) PatchEdge(ctx context.Context, edge *apipb.Patch) (*apipb.Edge, error) {
	return g.runtime.PatchEdge(ctx, edge)
}

func (g *Graph) PatchEdges(ctx context.Context, edges *apipb.Patches) (*apipb.Edges, error) {
	return g.runtime.PatchEdges(ctx, edges)
}

func (g *Graph) DelEdge(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	return g.runtime.DelEdge(ctx, path)
}

func (g *Graph) DelEdges(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	return g.runtime.DelEdges(ctx, paths)
}

func (g *Graph) Publish(ctx context.Context, message *apipb.OutboundMessage) (*empty.Empty, error) {
	return &empty.Empty{}, g.runtime.Machine().PubSub().Publish(message.Channel, &apipb.Message{
		Channel:   message.Channel,
		Data:      message.Data,
		Sender:    g.runtime.NodeContext(ctx).Path,
		Timestamp: time.Now().UnixNano(),
	})
}

func (g *Graph) Subscribe(filter *apipb.ChannelFilter, server apipb.GraphService_SubscribeServer) error {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		if val, ok := msg.(apipb.Mapper); ok {
			result, err := vm.Eval(programs, val)
			if err != nil {
				logger.Error("subscription filter failure", zap.Error(err))
				return false
			}
			return result
		}
		return false
	}
	if err := g.runtime.Machine().PubSub().SubscribeFilter(server.Context(), filter.Channel, filterFunc, func(msg interface{}) {
		if err, ok := msg.(error); ok && err != nil {
			logger.Error("failed to send subscription", zap.Error(err))
			return
		}
		if val, ok := msg.(*apipb.Message); ok && val != nil {
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
	return g.runtime.EdgesFrom(ctx, filter)
}

func (g *Graph) EdgesTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	return g.runtime.EdgesTo(ctx, filter)
}

func (g *Graph) Import(ctx context.Context, graph *apipb.Graph) (*apipb.Graph, error) {
	return g.runtime.Import(ctx, graph)
}

func (g *Graph) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	return g.runtime.SubGraph(ctx, filter)
}

func (g *Graph) Export(ctx context.Context, empty *empty.Empty) (*apipb.Graph, error) {
	return g.runtime.Export(ctx)
}

func (g *Graph) GetSchema(ctx context.Context, empty *empty.Empty) (*apipb.Schema, error) {
	return g.runtime.Schema(ctx)
}

func (g *Graph) Shutdown(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	go g.runtime.Close()
	return &empty.Empty{}, nil
}
