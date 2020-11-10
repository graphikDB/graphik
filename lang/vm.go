package lang

import "github.com/autom8ter/graphik/graph"

type VM struct {
	g       *graph.Graph
	private FuncMap
	public  FuncMap
}

func NewVM(g *graph.Graph) *VM {
	return &VM{
		g:       g,
		private: privateFuncMap(g),
	}
}

func (v *VM) Graph() *graph.Graph {
	return v.g
}

func (v *VM) Private() FuncMap {
	return v.private
}

func (v *VM) Public() FuncMap {
	return v.public
}
