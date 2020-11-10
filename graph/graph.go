package graph

import (
	"sort"
	"time"
)

type Graph struct {
	nodes     map[string]map[string]Values
	edges     map[string]map[string]Values
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

func New() *Graph {
	return &Graph{
		nodes:     map[string]map[string]Values{},
		edges:     map[string]map[string]Values{},
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
	}
}

func (g *Graph) Do(fn func(g *Graph)) {
	fn(g)
}

func (n *Graph) NodeCount(nodeType string) int {
	if c, ok := n.nodes[nodeType]; ok {
		return len(c)
	}
	return 0
}

func (n *Graph) NodeTypes() []string {
	var nodeTypes []string
	for k, _ := range n.nodes {
		nodeTypes = append(nodeTypes, k)
	}
	sort.Strings(nodeTypes)
	return nodeTypes
}

func (n *Graph) AllNodes() ValueSet {
	var nodes ValueSet
	n.RangeNode(Any, func(node Values) bool {
		nodes = append(nodes, node)
		return true
	})
	nodes.Sort()
	return nodes
}

func (n *Graph) GetNode(path string) (Values, bool) {
	xtype, xid := SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		node := c[xid]
		return node, node != nil
	}
	return nil, false
}

func (n *Graph) SetNode(value Values) Values {
	if value.GetID() == "" {
		value.SetID(UUID())
	}
	if value.GetType() == "" {
		value.SetType(Default)
	}
	if value.GetCreatedAt() == 0 {
		value.SetCreatedAt(time.Now())
	}
	if value.GetUpdatedAt() == 0 {
		value.SetUpdatedAt(time.Now())
	}
	if _, ok := n.nodes[value.GetType()]; !ok {
		n.nodes[value.GetType()] = map[string]Values{}
	}
	n.nodes[value.GetType()][value.GetID()] = value
	return value
}

func (n *Graph) PatchNode(value Values) Values {
	value.SetUpdatedAt(time.Now())
	if _, ok := n.nodes[value.GetType()]; !ok {
		return nil
	}
	node := n.nodes[value.GetType()][value.GetID()]
	for k, v := range value {
		node[k] = v
	}
	return node
}

func (n *Graph) RangeNode(nodeType string, f func(node Values) bool) {
	if nodeType == Any {
		for _, c := range n.nodes {
			for _, node := range c {
				if !f(node) {
					return
				}
			}
		}
	} else {
		if c, ok := n.nodes[nodeType]; ok {
			for _, node := range c {
				if !f(node) {
					return
				}
			}
		}
	}
}

func (n *Graph) DeleteNode(path string) bool {
	if !n.HasNode(path) {
		return false
	}
	n.RangeFrom(path, func(e Values) bool {
		n.DeleteEdge(e.GetPath())
		if this := e.GetString(CascadeKey); this != "" {
			if this == CascadeTo || this == CascadeMutual {
				n.DeleteEdge(e.GetString(ToKey))
			}
		}
		return true
	})
	n.RangeTo(path, func(e Values) bool {
		n.DeleteEdge(e.GetPath())
		if this := e.GetString(CascadeKey); this != "" {
			if this == CascadeFrom || this == CascadeMutual {
				n.DeleteNode(e.GetString(FromKey))
			}
		}
		return true
	})
	xtype, xid := SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		delete(c, xid)
	}
	return true
}

func (n *Graph) HasNode(path string) bool {
	_, ok := n.GetNode(path)
	return ok
}

func (n *Graph) FilterNode(nodeType string, filter func(node Values) bool) ValueSet {
	var filtered ValueSet
	n.RangeNode(nodeType, func(node Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	filtered.Sort()
	return filtered
}

func (n *Graph) SetNodes(nodes ValueSet) {
	for _, node := range nodes {
		n.SetNode(node)
	}
}

func (n *Graph) DeleteNodes(nodes ValueSet) {
	for _, node := range nodes {
		n.DeleteNode(node.GetPath())
	}
}

func (n *Graph) ClearNodes(nodeType string) {
	if cache, ok := n.nodes[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n *Graph) FilterSearchNodes(filter *Filter) (ValueSet, error) {
	var nodes ValueSet
	var err error
	var pass bool
	n.RangeNode(filter.Type, func(node Values) bool {
		pass, err = BooleanExpression(filter.Expressions, node)
		if err != nil {
			return false
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})
	nodes.Sort()
	return nodes, err
}

func (n *Graph) EdgeCount(edgeType string) int {
	if c, ok := n.edges[edgeType]; ok {
		return len(c)
	}
	return 0
}

func (n *Graph) EdgeTypes() []string {
	var edgeTypes []string
	for k, _ := range n.edges {
		edgeTypes = append(edgeTypes, k)
	}
	sort.Strings(edgeTypes)
	return edgeTypes
}

func (n *Graph) AllEdges() ValueSet {
	var edges ValueSet
	n.RangeEdges(Any, func(edge Values) bool {
		edges = append(edges, edge)
		return true
	})
	edges.Sort()
	return edges
}

func (n *Graph) GetEdge(path string) (Values, bool) {
	typ, id := SplitPath(path)
	if c, ok := n.edges[typ]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n *Graph) SetEdge(value Values) Values {
	if value.GetID() == "" {
		value.SetID(UUID())
	}
	if value.GetType() == "" {
		value.SetType(Default)
	}
	if _, ok := n.edges[value.GetType()]; !ok {
		n.edges[value.GetType()] = map[string]Values{}
	}

	n.edges[value.GetType()][value.GetID()] = value

	n.edgesFrom[value.GetString(FromKey)] = append(n.edgesFrom[value.GetString(FromKey)], value.GetPath())
	n.edgesTo[value.GetString(ToKey)] = append(n.edgesTo[value.GetString(ToKey)], value.GetPath())

	if value.GetBool(MutualKey) {
		n.edgesTo[value.GetString(FromKey)] = append(n.edgesTo[value.GetString(FromKey)], value.GetPath())
		n.edgesFrom[value.GetString(ToKey)] = append(n.edgesFrom[value.GetString(ToKey)], value.GetPath())
	}
	return value
}

func (n *Graph) RangeEdges(edgeType string, f func(edge Values) bool) {
	if edgeType == Any {
		for _, c := range n.edges {
			for _, v := range c {
				if !f(v) {
					return
				}
			}
		}
	} else {
		if c, ok := n.edges[edgeType]; ok {
			for _, v := range c {
				if !f(v) {
					return
				}
			}
		}
	}
}

func (n *Graph) DeleteEdge(path string) {
	xtype, xid := SplitPath(path)
	edge, ok := n.GetEdge(path)
	if !ok {
		return
	}
	n.edgesFrom[edge.GetString(FromKey)] = removeEdge(path, n.edgesFrom[edge.GetString(FromKey)])
	n.edgesTo[edge.GetString(FromKey)] = removeEdge(path, n.edgesTo[edge.GetString(FromKey)])
	n.edgesFrom[edge.GetString(ToKey)] = removeEdge(path, n.edgesFrom[edge.GetString(ToKey)])
	n.edgesTo[edge.GetString(ToKey)] = removeEdge(path, n.edgesTo[edge.GetString(ToKey)])
	delete(n.edges[xtype], xid)
}

func (n *Graph) HasEdge(path string) bool {
	_, ok := n.GetEdge(path)
	return ok
}

func (e *Graph) RangeFrom(path string, fn func(e Values) bool) {
	for _, edge := range e.edgesFrom[path] {
		e, ok := e.GetEdge(edge)
		if ok {
			if !fn(e) {
				return
			}
		}
	}
}

func (e *Graph) RangeTo(path string, fn func(e Values) bool) {
	for _, edge := range e.edgesTo[path] {
		e, ok := e.GetEdge(edge)
		if ok {
			if !fn(e) {
				return
			}
		}
	}
}

func (e *Graph) RangeFilterFrom(path string, filter *Filter) ValueSet {
	var edges ValueSet
	e.RangeFrom(path, func(e Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	edges.Sort()
	return edges
}

func (e *Graph) RangeFilterTo(path string, filter *Filter) ValueSet {
	var edges ValueSet
	e.RangeTo(path, func(e Values) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	edges.Sort()
	return edges
}

func (n *Graph) FilterEdges(edgeType string, filter func(edge Values) bool) ValueSet {
	var filtered ValueSet
	n.RangeEdges(edgeType, func(node Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	filtered.Sort()
	return filtered
}

func (n *Graph) SetEdges(edges ValueSet) {
	for _, edge := range edges {
		n.SetEdge(edge)
	}
}

func (n *Graph) DeleteEdges(edges ValueSet) {
	for _, edge := range edges {
		n.DeleteEdge(edge.GetPath())
	}
}

func (n *Graph) ClearEdges(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.DeleteEdge(v.GetPath())
		}
	}
}

func (e *Graph) PatchEdge(value Values) Values {
	if _, ok := e.edges[value.GetType()]; !ok {
		return nil
	}
	for k, v := range value {
		e.edges[value.GetType()][value.GetID()][k] = v
	}
	e.edges[value.GetType()][value.GetID()].SetUpdatedAt(time.Now())
	return e.edges[value.GetType()][value.GetID()]
}

func (e *Graph) FilterSearchEdges(filter *Filter) (ValueSet, error) {
	var edges ValueSet
	var err error
	var pass bool
	e.RangeEdges(filter.Type, func(edge Values) bool {
		pass, _ = BooleanExpression(filter.Expressions, edge)
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < filter.Limit
	})
	edges.Sort()
	return edges, err
}

func removeEdge(path string, paths []string) []string {
	var newPaths []string
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	sort.Strings(newPaths)
	return newPaths
}
