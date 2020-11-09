package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
)

type EdgeStore struct {
	edges     map[string]map[string]*lang.Values
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

func newEdgeStore() *EdgeStore {
	return &EdgeStore{
		edges:     map[string]map[string]*lang.Values{},
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
	}
}

func (n *EdgeStore) Len(edgeType string) int {
	if c, ok := n.edges[edgeType]; ok {
		return len(c)
	}
	return 0
}

func (n *EdgeStore) Types() []string {
	var edgeTypes []string
	for k, _ := range n.edges {
		edgeTypes = append(edgeTypes, k)
	}
	return edgeTypes
}

func (n *EdgeStore) All() []*lang.Values {
	var edges []*lang.Values
	n.Range(apipb.Keyword_ANY.String(), func(edge *lang.Values) bool {
		edges = append(edges, edge)
		return true
	})
	return edges
}

func (n *EdgeStore) Get(path string) (*lang.Values, bool) {
	typ, id := lang.SplitPath(path)
	if c, ok := n.edges[typ]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n *EdgeStore) Set(value *lang.Values) *lang.Values {
	if _, ok := n.edges[value.GetType()]; !ok {
		n.edges[value.GetType()] = map[string]*lang.Values{}
	}

	n.edges[value.GetType()][value.GetID()] = value

	n.edgesFrom[value.GetString("from")] = append(n.edgesFrom[value.GetString("from")], value.PathString())
	n.edgesTo[value.GetString("to")] = append(n.edgesTo[value.GetString("to")], value.PathString())

	if value.GetBool("mutual") {
		n.edgesTo[value.GetString("from")] = append(n.edgesTo[value.GetString("from")], value.PathString())
		n.edgesFrom[value.GetString("to")] = append(n.edgesFrom[value.GetString("to")], value.PathString())
	}
	return value
}

func (n *EdgeStore) Range(edgeType string, f func(edge *lang.Values) bool) {
	if edgeType == apipb.Keyword_ANY.String() {
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

func (n *EdgeStore) Delete(path string) {
	xtype, xid := lang.SplitPath(path)
	edge, ok := n.Get(path)
	if !ok {
		return
	}
	n.edgesFrom[edge.GetString("from")] = removeEdge(path, n.edgesFrom[edge.GetString("from")])
	n.edgesTo[edge.GetString("from")] = removeEdge(path, n.edgesTo[edge.GetString("from")])
	n.edgesFrom[edge.GetString("to")] = removeEdge(path, n.edgesFrom[edge.GetString("to")])
	n.edgesTo[edge.GetString("to")] = removeEdge(path, n.edgesTo[edge.GetString("to")])
	delete(n.edges[xtype], xid)
}

func (n *EdgeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (e *EdgeStore) RangeFrom(path string, fn func(e *lang.Values) bool) {
	for _, edge := range e.edgesFrom[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeTo(path string, fn func(e *lang.Values) bool) {
	for _, edge := range e.edgesTo[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeFilterFrom(path string, filter *apipb.Filter) []*lang.Values {
	var edges []*lang.Values
	e.RangeFrom(path, func(e *lang.Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := lang.BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (e *EdgeStore) RangeFilterTo(path string, filter *apipb.Filter) []*lang.Values {
	var edges []*lang.Values
	e.RangeTo(path, func(e *lang.Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := lang.BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return edges
}

func (n *EdgeStore) Filter(edgeType string, filter func(edge *lang.Values) bool) []*lang.Values {
	var filtered []*lang.Values
	n.Range(edgeType, func(node *lang.Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n *EdgeStore) SetAll(edges []*lang.Values) {
	for _, edge := range edges {
		n.Set(edge)
	}
}

func (n *EdgeStore) DeleteAll(edges []*lang.Values) {
	for _, edge := range edges {
		n.Delete(edge.PathString())
	}
}

func (n *EdgeStore) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(v.PathString())
		}
	}
}

func (n *EdgeStore) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func (e *EdgeStore) Patch(updatedAt int64, value *lang.Values) *lang.Values {
	if _, ok := e.edges[value.GetType()]; !ok {
		return nil
	}
	for k, v := range value.Fields {
		e.edges[value.GetType()][value.GetID()].Fields[k] = v
	}
	e.edges[value.GetType()][value.GetID()].Fields["updated_at"] = lang.ToValue(updatedAt)
	return e.edges[value.GetType()][value.GetID()]
}

func (e *EdgeStore) FilterSearch(filter *apipb.Filter) ([]*lang.Values, error) {
	var edges []*lang.Values
	var err error
	var pass bool
	e.Range(filter.Type, func(edge *lang.Values) bool {
		pass, err = lang.BooleanExpression(filter.Expressions, edge)
		if err != nil {
			return false
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	})
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
