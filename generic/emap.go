package generic

import "github.com/autom8ter/graphik/graph/model"

type EdgeMap struct {
	edges map[string]map[string]*model.Edge
	edgesTo map[string]map[string][]*model.Edge
	edgesFrom map[string]map[string][]*model.Edge
}

func (n EdgeMap) Len(edgeType string) int {
	if c, ok := n.edges[edgeType]; ok {
		return len(c)
	}
	return 0
}

func (n EdgeMap) Types() []string {
	var edgeTypes []string
	for k, _ := range n.edges {
		edgeTypes = append(edgeTypes, k)
	}
	return edgeTypes
}

func (n EdgeMap) Get(edgeType string, id string) (*model.Edge, bool) {
	if c, ok := n.edges[edgeType]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n EdgeMap) Set(value *model.Edge) {
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

func (n EdgeMap) Range(edgeType string, f func(node *model.Edge) bool) {
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

func (n EdgeMap) Delete(edgeType string, id string) {
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

func (n EdgeMap) Exists(edgeType string, id string) bool {
	_, ok := n.Get(edgeType, id)
	return ok
}

func (n EdgeMap) Filter(edgeType string, filter func(node *model.Edge) bool) []*model.Edge {
	var filtered []*model.Edge
	n.Range(edgeType, func(node *model.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n EdgeMap) SetAll(edges ...*model.Edge) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n EdgeMap) Clear(edgeType string) {
	if cache, ok := n[edgeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n EdgeMap) Close() {
	for edgeType, _ := range n {
		n.Clear(edgeType)
	}
}

func removeEdge(id string, edges []*model.Edge) []*model.Edge{
	var newEdges []*model.Edge
	for _, edge := range edges {
		if edge.ID != id {
			newEdges = append(newEdges, edge)
		}
	}
	return newEdges
}
