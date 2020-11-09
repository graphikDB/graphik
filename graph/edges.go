package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type EdgeStore struct {
	edges     map[string]map[string]*structpb.Struct
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

func newEdgeStore() *EdgeStore {
	return &EdgeStore{
		edges:     map[string]map[string]*structpb.Struct{},
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

func (n *EdgeStore) All() *structpb.Struct {
	var edges []*structpb.Struct
	n.Range(apipb.Keyword_ANY.String(), func(edge *structpb.Struct) bool {
		edges = append(edges, edge)
		return true
	})
	return &apipb.Valuess{
		Edges: edges,
	}
}

func (n *EdgeStore) Get(path string) (*structpb.Struct, bool) {
	typ, id := lang.SplitPath(path)
	if c, ok := n.edges[typ]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n *EdgeStore) Set(value *structpb.Struct) *structpb.Struct {
	if _, ok := n.edges[lang.GetType(value)]; !ok {
		n.edges[lang.GetType(value)] = map[string]*structpb.Struct{}
	}

	n.edges[lang.GetType(value)][value.ID()] = value

	n.edgesFrom[value.From] = append(n.edgesFrom[value.From], value.Path)
	n.edgesTo[value.To] = append(n.edgesTo[value.To], value.Path)

	if value.Mutual {
		n.edgesTo[value.From] = append(n.edgesTo[value.From], value.Path)
		n.edgesFrom[value.To] = append(n.edgesFrom[value.To], value.Path)
	}
	return value
}

func (n *EdgeStore) Range(edgeType string, f func(edge *structpb.Struct) bool) {
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
	n.edgesFrom[edge.From] = removeEdge(edge.Path, n.edgesFrom[edge.From])
	n.edgesTo[edge.From] = removeEdge(edge.Path, n.edgesTo[edge.From])
	n.edgesFrom[edge.To] = removeEdge(edge.Path, n.edgesFrom[edge.To])
	n.edgesTo[edge.To] = removeEdge(edge.Path, n.edgesTo[edge.To])
	delete(n.edges[xtype], xid)
}

func (n *EdgeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (e *EdgeStore) RangeFrom(path string, fn func(e *structpb.Struct) bool) {
	for _, edge := range e.edgesFrom[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeTo(path string, fn func(e *structpb.Struct) bool) {
	for _, edge := range e.edgesTo[path] {
		e, ok := e.Get(edge)
		if ok {
			if !fn(e) {
				break
			}
		}
	}
}

func (e *EdgeStore) RangeFilterFrom(path string, filter *apipb.Filter) *structpb.Struct {
	var edges []*structpb.Struct
	e.RangeFrom(path, func(e *structpb.Struct) bool {
		xtype, _ := lang.SplitPath(e.Path)
		if xtype != filter.Type {
			return true
		}
		pass, _ := lang.BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Valuess{
		Edges: edges,
	}
}

func (e *EdgeStore) RangeFilterTo(path string, filter *apipb.Filter) *structpb.Struct {
	var edges []*structpb.Struct
	e.RangeTo(path, func(e *structpb.Struct) bool {
		xtype, _ := lang.SplitPath(e.Path)
		if xtype != filter.Type {
			return true
		}
		pass, _ := lang.BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Valuess{
		Edges: edges,
	}
}

func (n *EdgeStore) Filter(edgeType string, filter func(edge *structpb.Struct) bool) *structpb.Struct {
	var filtered []*structpb.Struct
	n.Range(edgeType, func(node *structpb.Struct) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return &apipb.Valuess{
		Edges: filtered,
	}
}

func (n *EdgeStore) SetAll(edges *structpb.Struct) {
	for _, edge := range edges.Edges {
		n.Set(edge)
	}
}

func (n *EdgeStore) DeleteAll(edges *structpb.Struct) {
	for _, edge := range edges.Edges {
		n.Delete(edge.Path)
	}
}

func (n *EdgeStore) Clear(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.Delete(v.Path)
		}
	}
}

func (n *EdgeStore) Close() {
	for _, edgeType := range n.Types() {
		n.Clear(edgeType)
	}
}

func (e *EdgeStore) Patch(updatedAt *timestamp.Timestamp, value *apipb.Patch) *structpb.Struct {
	if _, ok := e.edges[xtype]; !ok {
		return nil
	}
	for k, v := range value.Patch.Fields {
		e.edges[xtype][xid].Attributes.Fields[k] = v
	}
	e.edges[xtype][xid].UpdatedAt = updatedAt
	return e.edges[xtype][xid]
}

func (e *EdgeStore) FilterSearch(filter *apipb.Filter) (*structpb.Struct, error) {
	var edges []*structpb.Struct
	var err error
	var pass bool
	e.Range(filter.Type, func(edge *structpb.Struct) bool {
		pass, err = lang.BooleanExpression(filter.Expressions, edge)
		if err != nil {
			return false
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Valuess{
		Edges: edges,
	}, err
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
