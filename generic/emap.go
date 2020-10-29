package generic

import (
	"github.com/autom8ter/graphik/graph/model"
)

type Edges struct {
	edges     map[string]map[string]*model.Edge
	edgesTo   map[string]map[string][]*model.Edge
	edgesFrom map[string]map[string][]*model.Edge
}

func NewEdges() *Edges {
	return &Edges{
		edges: map[string]map[string]*model.Edge{},
		edgesTo: map[string]map[string][]*model.Edge{},
		edgesFrom: map[string]map[string][]*model.Edge{},
	}
}

func (n *Edges) Len(edgeType string) int {
	if c, ok := n.edges[edgeType]; ok {
		return len(c)
	}
	return 0
}

func (n *Edges) Types() []string {
	var edgeTypes []string
	for k, _ := range n.edges {
		edgeTypes = append(edgeTypes, k)
	}
	return edgeTypes
}

func (n *Edges) All() []*model.Edge {
	var edges []*model.Edge
	n.Range(Any, func(edge *model.Edge) bool {
		edges = append(edges, edge)
		return true
	})
	return edges
}

func (n *Edges) Get(key *model.ForeignKey) (*model.Edge, bool) {
	if c, ok := n.edges[key.Type]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n *Edges) Set(value *model.Edge) {
	if value.ID == "" {
		value.ID = uuid()
	}
	if _, ok := n.edges[value.Type]; !ok {
		n.edges[value.Type] = map[string]*model.Edge{}
	}
	if _, ok := n.edges[value.From.Type]; !ok {
		n.edgesFrom[value.From.Type] = map[string][]*model.Edge{}
	}
	if _, ok := n.edges[value.To.Type]; !ok {
		n.edgesTo[value.To.Type] = map[string][]*model.Edge{}
	}
	n.edges[value.Type][value.ID] = value
	n.edgesFrom[value.From.Type][value.From.ID] = append(n.edgesFrom[value.From.Type][value.From.ID], value)
	n.edgesTo[value.To.Type][value.To.ID] = append(n.edgesTo[value.To.Type][value.To.ID], value)
}

func (n *Edges) Range(edgeType string, f func(edge *model.Edge) bool) {
	if edgeType == Any {
		for _, c := range n.edges {
			for _, v := range c {
				f(v)
			}
		}
	} else {
		if c, ok := n.edges[edgeType]; ok {
			for _, v := range c {
				f(v)
			}
		}
	}
}

func (n *Edges) Delete(edgeType string, id string) {
	edge, ok := n.Get(edgeType, id)
	if !ok {
		return
	}
	if edges, ok := n.edgesFrom[edge.From.Type][edge.From.ID]; ok {
		edges = removeEdge(edge.ID, edges)
		n.edgesFrom[edge.From.Type][edge.From.ID] = edges
	}
	if edges, ok := n.edgesTo[edge.To.Type][edge.To.ID]; ok {
		edges = removeEdge(edge.ID, edges)
		n.edgesTo[edge.To.Type][edge.To.ID] = edges
	}
	if c, ok := n.edges[edgeType]; ok {
		delete(c, id)
	}
}

func (n *Edges) Exists(edgeType string, id string) bool {
	_, ok := n.Get(edgeType, id)
	return ok
}

func (e Edges) RangeFrom(node *model.Node, fn func(e *model.Edge) bool) {
	if _, ok := e.edgesFrom[node.Type]; !ok {
		return
	}
	if _, ok := e.edgesFrom[node.Type][node.ID]; !ok {
		return
	}
	for _, edge := range e.edgesFrom[node.Type][node.ID] {
		if !fn(edge) {
			break
		}
	}
}

func (e Edges) RangeTo(node *model.Node, fn func(e *model.Edge) bool) {
	if _, ok := e.edgesTo[node.Type]; !ok {
		return
	}
	if _, ok := e.edgesTo[node.Type][node.ID]; !ok {
		return
	}
	for _, edge := range e.edgesTo[node.Type][node.ID] {
		if !fn(edge) {
			break
		}
	}
}

func (e Edges) EdgesFrom(node *model.Node) []*model.Edge {
	if _, ok := e.edgesFrom[node.Type]; !ok {
		return nil
	}
	if _, ok := e.edgesFrom[node.Type][node.ID]; !ok {
		return nil
	}
	return e.edgesFrom[node.Type][node.ID]
}

func (e Edges) EdgesTo(nodeType, nodeID string) []*model.Edge {
	if _, ok := e.edgesTo[nodeType]; !ok {
		return nil
	}
	if _, ok := e.edgesTo[nodeType][nodeID]; !ok {
		return nil
	}
	return e.edgesTo[nodeType][nodeID]
}

func (n *Edges) Filter(edgeType string, filter func(edge *model.Edge) bool) []*model.Edge {
	var filtered []*model.Edge
	n.Range(edgeType, func(node *model.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n *Edges) SetAll(edges ...*model.Edge) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n *Edges) DeleteAll(edges ...*model.Edge) {
	for _, edge := range edges {
		n.Delete(edge.Type, edge.ID)
	}
}

func (n *Edges) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(v.Type, v.ID)
		}
	}
}

func (n *Edges) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func removeEdge(id string, edges []*model.Edge) []*model.Edge {
	var newEdges []*model.Edge
	for _, edge := range edges {
		if edge.ID != id {
			newEdges = append(newEdges, edge)
		}
	}
	return newEdges
}
