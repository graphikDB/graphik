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
