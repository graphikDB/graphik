package generic

import (
	"github.com/autom8ter/graphik/graph/model"
	"github.com/jmespath/go-jmespath"
	"time"
)

type Edges struct {
	edges     map[string]map[string]*model.Edge
	edgesTo   map[string]map[string][]*model.Edge
	edgesFrom map[string]map[string][]*model.Edge
}

func NewEdges() *Edges {
	return &Edges{
		edges:     map[string]map[string]*model.Edge{},
		edgesTo:   map[string]map[string][]*model.Edge{},
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
	EdgeList(edges).Sort()
	return edges
}

func (n *Edges) Get(key model.ForeignKey) (*model.Edge, bool) {
	if c, ok := n.edges[key.Type]; ok {
		node := c[key.ID]
		return c[key.ID], node != nil
	}
	return nil, false
}

func (n *Edges) set(value *model.Edge) *model.Edge {
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
	return value
}

func (n *Edges) Set(value *model.Edge) *model.Edge {
	e := n.set(value)
	if value.Mutual != nil && *value.Mutual {
		n.set(&model.Edge{
			Type:       value.Type,
			Attributes: value.Attributes,
			From:       value.To,
			To:         value.From,
			Mutual:     value.Mutual,
			CreatedAt:  e.CreatedAt,
			UpdatedAt:  e.UpdatedAt,
		})
	}
	return e
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

func (n *Edges) Delete(key model.ForeignKey) {
	edge, ok := n.Get(key)
	if !ok {
		return
	}
	if edges, ok := n.edgesFrom[edge.From.Type][edge.From.ID]; ok {
		n.edgesFrom[edge.From.Type][edge.From.ID] = removeEdge(edge.ID, edges)
	}
	if edges, ok := n.edgesTo[edge.To.Type][edge.To.ID]; ok {
		n.edgesTo[edge.To.Type][edge.To.ID] = removeEdge(edge.ID, edges)
	}
	if c, ok := n.edges[key.Type]; ok {
		delete(c, key.ID)
	}
}

func (n *Edges) Exists(key model.ForeignKey) bool {
	_, ok := n.Get(key)
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
	edges := e.edgesFrom[node.Type][node.ID]
	EdgeList(edges).Sort()
	return edges
}

func (e Edges) EdgesTo(nodeType, nodeID string) []*model.Edge {
	if _, ok := e.edgesTo[nodeType]; !ok {
		return nil
	}
	if _, ok := e.edgesTo[nodeType][nodeID]; !ok {
		return nil
	}
	edges := e.edgesTo[nodeType][nodeID]
	EdgeList(edges).Sort()
	return edges
}

func (n *Edges) Filter(edgeType string, filter func(edge *model.Edge) bool) []*model.Edge {
	var filtered []*model.Edge
	n.Range(edgeType, func(node *model.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	EdgeList(filtered).Sort()
	return filtered
}

func (n *Edges) SetAll(edges ...*model.Edge) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n *Edges) DeleteAll(edges ...*model.Edge) {
	for _, edge := range edges {
		n.Delete(model.ForeignKey{
			ID:   edge.ID,
			Type: edge.Type,
		})
	}
}

func (n *Edges) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(model.ForeignKey{
				ID:   v.ID,
				Type: v.Type,
			})
		}
	}
}

func (n *Edges) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func (e *Edges) Patch(updatedAt time.Time, value *model.Patch) *model.Edge {
	if _, ok := e.edges[value.Type]; !ok {
		return nil
	}
	for k, v := range value.Patch {
		e.edges[value.Type][value.ID].Attributes[k] = v
	}
	e.edges[value.Type][value.ID].UpdatedAt = &updatedAt
	return e.edges[value.Type][value.ID]
}

func (e *Edges) Search(expression, nodeType string) (*model.SearchResults, error) {
	results := &model.SearchResults{
		Search: expression,
	}
	exp, err := jmespath.Compile(expression)
	if err != nil {
		return nil, err
	}
	e.Range(nodeType, func(edge *model.Edge) bool {
		val, _ := exp.Search(edge)
		if val != nil {
			results.Results = append(results.Results, &model.SearchResult{
				ID:   edge.ID,
				Type: edge.Type,
				Val:  val,
			})
		}
		return true
	})
	return results, nil
}

func (e *Edges) FilterSearch(filter model.Filter) ([]*model.Edge, error) {
	var edges []*model.Edge
	e.Range(filter.Type, func(edge *model.Edge) bool {
		for _, exp := range filter.Statements {
			val, _ := jmespath.Search(exp.Expression, edge)
			if exp.Operator == model.OperatorNeq {
				if val == exp.Value {
					return true
				}
			}
			if exp.Operator == model.OperatorEq {
				if val != exp.Value {
					return true
				}
			}
		}
		edges = append(edges, edge)
		return len(edges) < filter.Limit
	})
	EdgeList(edges).Sort()
	return edges, nil
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


type EdgeList []*model.Edge

func (n EdgeList) Sort() {
	sorter := Interface{
		LenFunc: func() int {
			if n == nil {
				return 0
			}
			return len(n)
		},
		LessFunc: func(i, j int) bool {
			if n == nil {
				return false
			}
			return n[i].UpdatedAt.UnixNano() > n[j].UpdatedAt.UnixNano()
		},
		SwapFunc: func(i, j int) {
			if n == nil {
				return
			}
			n[i], n[j] = n[j], n[i]
		},
	}
	sorter.Sort()
}