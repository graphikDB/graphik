package service

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/runtime"
)

type Graph struct {
	runtime *runtime.Runtime
}

func NewGraph(runtime *runtime.Runtime) *Graph {
	return &Graph{runtime: runtime}
}

func (g Graph) CreateNode(ctx context.Context, node *apipb.Node) (*apipb.Node, error) {
	panic("implement me")
}

func (g Graph) CreateNodes(ctx context.Context, nodes *apipb.Nodes) (*apipb.Nodes, error) {
	panic("implement me")
}

func (g Graph) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	panic("implement me")
}

func (g Graph) GetNodes(ctx context.Context, paths *apipb.Paths) (*apipb.Nodes, error) {
	panic("implement me")
}

func (g Graph) SearchNodes(ctx context.Context, filter *apipb.TypeFilter) (*apipb.Nodes, error) {
	panic("implement me")
}

func (g Graph) PatchNode(ctx context.Context, node *apipb.Node) (*apipb.Node, error) {
	panic("implement me")
}

func (g Graph) PatchNodes(ctx context.Context, nodes *apipb.Nodes) (*apipb.Nodes, error) {
	panic("implement me")
}

func (g Graph) DelNode(ctx context.Context, path *apipb.Path) (*apipb.Counter, error) {
	panic("implement me")
}

func (g Graph) DelNodes(ctx context.Context, paths *apipb.Paths) (*apipb.Counter, error) {
	panic("implement me")
}

func (g Graph) CreateEdge(ctx context.Context, edge *apipb.Edge) (*apipb.Edge, error) {
	panic("implement me")
}

func (g Graph) CreateEdges(ctx context.Context, edges *apipb.Edges) (*apipb.Edges, error) {
	panic("implement me")
}

func (g Graph) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	panic("implement me")
}

func (g Graph) GetEdges(ctx context.Context, paths *apipb.Paths) (*apipb.Edges, error) {
	panic("implement me")
}

func (g Graph) SearchEdges(ctx context.Context, filter *apipb.TypeFilter) (*apipb.Edges, error) {
	panic("implement me")
}

func (g Graph) PatchEdge(ctx context.Context, edge *apipb.Edge) (*apipb.Edge, error) {
	panic("implement me")
}

func (g Graph) PatchEdges(ctx context.Context, edges *apipb.Edges) (*apipb.Edges, error) {
	panic("implement me")
}

func (g Graph) DelEdge(ctx context.Context, path *apipb.Path) (*apipb.Counter, error) {
	panic("implement me")
}

func (g Graph) DelEdges(ctx context.Context, paths *apipb.Paths) (*apipb.Counter, error) {
	panic("implement me")
}
