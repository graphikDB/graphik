package generic

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

type Edges struct {
	edges     map[string]map[string]*apipb.Edge
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

func NewEdges() *Edges {
	return &Edges{
		edges:     map[string]map[string]*apipb.Edge{},
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
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

func (n *Edges) All() []*apipb.Edge {
	var edges []*apipb.Edge
	n.Range(Any, func(edge *apipb.Edge) bool {
		edges = append(edges, edge)
		return true
	})
	EdgeList(edges).Sort()
	return edges
}

func (n *Edges) Get(path *apipb.Path) (*apipb.Edge, bool) {
	if c, ok := n.edges[path.Type]; ok {
		node := c[path.ID]
		return c[path.ID], node != nil
	}
	return nil, false
}

func (n *Edges) Set(value *apipb.Edge) *apipb.Edge {
	if _, ok := n.edges[value.Path.Type]; !ok {
		n.edges[value.Path.Type] = map[string]*apipb.Edge{}
	}

	n.edges[value.Path.Type][value.Path.ID] = value

	path := value.Path.String()

	n.edgesFrom[value.From.String()] = append(n.edgesFrom[value.From.String()], path)
	n.edgesTo[value.To.String()] = append(n.edgesTo[value.To.String()], path)

	if value.Mutual {
		n.edgesTo[value.From.String()] = append(n.edgesTo[value.From.String()], path)
		n.edgesFrom[value.To.String()] = append(n.edgesFrom[value.To.String()], path)
	}
	return value
}

func (n *Edges) Range(edgeType string, f func(edge *apipb.Edge) bool) {
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

func (n *Edges) Delete(path *apipb.Path) {
	edge, ok := n.Get(path)
	if !ok {
		return
	}
	n.edgesFrom[edge.From.String()] = removeEdge(edge.Path.String(), n.edgesFrom[edge.From.String()])
	n.edgesTo[edge.From.String()] = removeEdge(edge.Path.String(), n.edgesTo[edge.From.String()])
	n.edgesFrom[edge.To.String()] = removeEdge(edge.Path.String(), n.edgesFrom[edge.To.String()])
	n.edgesTo[edge.To.String()] = removeEdge(edge.Path.String(), n.edgesTo[edge.To.String()])
	delete(n.edges[path.Type], path.ID)
}

func (n *Edges) Exists(path *apipb.Path) bool {
	_, ok := n.Get(path)
	return ok
}

func (e Edges) RangeFrom(path *apipb.Path, fn func(e *apipb.Edge) bool) {
	for _, edge := range e.edgesFrom[path.String()] {
		e, ok := e.Get(apipb.PathFromString(edge))
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e Edges) RangeTo(path *apipb.Path, fn func(e *apipb.Edge) bool) {
	for _, edge := range e.edgesTo[path.String()] {
		e, ok := e.Get(apipb.PathFromString(edge))
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e Edges) RangeFilterFrom(path *apipb.Path, filter apipb.Filter) []*apipb.Edge {
	var edges []*apipb.Edge
	e.RangeFrom(path, func(e *apipb.Edge) bool {
		if e.Path.Type != filter.Type {
			return true
		}
		pass, _ := apipb.Evaluate(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (e Edges) RangeFilterTo(path *apipb.Path, filter apipb.Filter) []*apipb.Edge {
	var edges []*apipb.Edge
	e.RangeTo(path, func(e *apipb.Edge) bool {
		if e.Path.Type != filter.Type {
			return true
		}
		pass, _ := apipb.Evaluate(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (n *Edges) Filter(edgeType string, filter func(edge *apipb.Edge) bool) []*apipb.Edge {
	var filtered []*apipb.Edge
	n.Range(edgeType, func(node *apipb.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	EdgeList(filtered).Sort()
	return filtered
}

func (n *Edges) SetAll(edges ...*apipb.Edge) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n *Edges) DeleteAll(edges ...*apipb.Edge) {
	for _, edge := range edges {
		n.Delete(edge.Path)
	}
}

func (n *Edges) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(v.Path)
		}
	}
}

func (n *Edges) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func (e *Edges) Patch(updatedAt time.Time, value *apipb.Patch) *apipb.Edge {
	if _, ok := e.edges[value.Path.Type]; !ok {
		return nil
	}
	for k, v := range value.Patch.Fields {
		e.edges[value.Path.Type][value.Path.ID].Attributes.Fields[k] = v
	}
	e.edges[value.Path.Type][value.Path.ID].UpdatedAt = &timestamp.Timestamp{
		Seconds:              updatedAt.Unix(),
	}
	return e.edges[value.Path.Type][value.Path.ID]
}

func (e *Edges) FilterSearch(filter apipb.Filter) ([]*apipb.Edge, error) {
	var edges []*apipb.Edge
	var err error
	var pass bool
	e.Range(filter.Type, func(edge *apipb.Edge) bool {
		pass, err = apipb.Evaluate(filter.Expressions, edge)
		if err != nil {
			return false
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	})
	EdgeList(edges).Sort()
	return edges, err
}

func removeEdge(path string, paths []string) []string {
	var newPaths []string
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	return newPaths
}

type EdgeList []*apipb.Edge

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
			return n[i].UpdatedAt.Nanos > n[j].UpdatedAt.Nanos
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
