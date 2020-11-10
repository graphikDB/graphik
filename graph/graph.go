package graph

type Graph struct {
	nodes *NodeStore
	edges *EdgeStore
}

func New() *Graph {
	edges := newEdgeStore()
	return &Graph{
		nodes: newNodeStore(edges),
		edges: edges,
	}
}

func (g *Graph) Edges() *EdgeStore {
	return g.edges
}

func (g *Graph) Nodes() *NodeStore {
	return g.nodes
}

func (g *Graph) Close() {
	g.edges.Close()
	g.nodes.Close()
}

func (g *Graph) Do(fn func(g *Graph)) {
	fn(g)
}
