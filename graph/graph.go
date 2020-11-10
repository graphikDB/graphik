package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"sort"
	"time"
)

type Graph struct {
	nodes     map[string]map[string]*apipb.Node
	edges     map[string]map[string]*apipb.Edge
	edgesTo   map[*apipb.Path][]*apipb.Path
	edgesFrom map[*apipb.Path][]*apipb.Path
}

func New() *Graph {
	return &Graph{
		nodes:     map[string]map[string]*apipb.Node{},
		edges:     map[string]map[string]*apipb.Edge{},
		edgesTo:   map[*apipb.Path][]*apipb.Path{},
		edgesFrom: map[*apipb.Path][]*apipb.Path{},
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

func (n *Graph) AllNodes() *apipb.Nodes {
	var nodes []*apipb.Node
	n.RangeNode(apipb.Keyword_ANY.String(), func(node *apipb.Node) bool {
		nodes = append(nodes, node)
		return true
	})
	return &apipb.Nodes{
		Nodes: nodes,
	}
}

func (n *Graph) GetNode(path *apipb.Path) (*apipb.Node, bool) {
	if c, ok := n.nodes[path.Type]; ok {
		node := c[path.Id]
		return node, node != nil
	}
	return nil, false
}

func (n *Graph) SetNode(value *apipb.Node) *apipb.Node {
	if value.GetId() == "" {
		value.Id = UUID()
	}
	if value.GetType() == "" {
		value.Type = apipb.Keyword_DEFAULT.String()
	}
	if value.GetCreatedAt() == 0 {
		value.CreatedAt = time.Now().UnixNano()
	}
	if value.GetUpdatedAt() == 0 {
		value.UpdatedAt = time.Now().UnixNano()
	}
	if _, ok := n.nodes[value.GetType()]; !ok {
		n.nodes[value.GetType()] = map[string]*apipb.Node{}
	}
	n.nodes[value.GetType()][value.GetId()] = value
	return value
}

func (n *Graph) PatchNode(value *apipb.Node) *apipb.Node {
	value.UpdatedAt = time.Now().UnixNano()
	if _, ok := n.nodes[value.GetType()]; !ok {
		return nil
	}
	node := n.nodes[value.GetType()][value.GetId()]
	for k, v := range value.GetAttributes().GetFields() {
		node.GetAttributes().GetFields()[k] = v
	}
	return node
}

func (n *Graph) PatchNodes(values []*apipb.Node) *apipb.Nodes {
	var nodes []*apipb.Node
	for _, val := range values {
		nodes = append(nodes, n.PatchNode(val))
	}
	return &apipb.Nodes{
		Nodes: nodes,
	}
}

func (n *Graph) RangeNode(nodeType string, f func(node *apipb.Node) bool) {
	if nodeType == apipb.Keyword_ANY.String() {
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

func (n *Graph) DeleteNode(path *apipb.Path) *apipb.Counter {
	if !n.HasNode(path) {
		return &apipb.Counter{
			Count: 0,
		}
	}
	n.RangeFrom(path, func(e *apipb.Edge) bool {
		n.DeleteEdge(e.Path())
		if e.Cascade == apipb.Cascade_CASCADE_TO || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			n.DeleteNode(e.To)
		}
		return true
	})
	n.RangeTo(path, func(e *apipb.Edge) bool {
		n.DeleteEdge(e.Path())
		if e.Cascade == apipb.Cascade_CASCADE_FROM || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			n.DeleteNode(e.From)
		}
		return true
	})

	delete(n.nodes[path.Type], path.Id)
	return &apipb.Counter{
		Count: 1,
	}
}

func (n *Graph) HasNode(path *apipb.Path) bool {
	_, ok := n.GetNode(path)
	return ok
}

func (n *Graph) FilterNode(nodeType string, filter func(node *apipb.Node) bool) *apipb.Nodes {
	var filtered []*apipb.Node
	n.RangeNode(nodeType, func(node *apipb.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return &apipb.Nodes{
		Nodes: filtered,
	}
}

func (n *Graph) SetNodes(nodes []*apipb.Node) *apipb.Nodes {
	var returned []*apipb.Node
	for _, node := range nodes {
		returned = append(returned, n.SetNode(node))
	}
	return &apipb.Nodes{
		Nodes: nodes,
	}
}

func (n *Graph) DeleteNodes(nodes []*apipb.Path) *apipb.Counter {
	count := int64(0)
	for _, node := range nodes {
		count += n.DeleteNode(node).Count
	}
	return &apipb.Counter{
		Count: count,
	}
}

func (n *Graph) ClearNodes(nodeType string) {
	if cache, ok := n.nodes[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n *Graph) FilterSearchNodes(filter *apipb.TypeFilter) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	var err error
	var pass bool
	n.RangeNode(filter.Type, func(node *apipb.Node) bool {
		pass, err = BooleanExpression(filter.Expressions, node)
		if err != nil {
			return false
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})

	return &apipb.Nodes{
		Nodes: nodes,
	}, err
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

func (n *Graph) AllEdges() *apipb.Edges {
	var edges []*apipb.Edge
	n.RangeEdges(apipb.Keyword_ANY.String(), func(edge *apipb.Edge) bool {
		edges = append(edges, edge)
		return true
	})
	return &apipb.Edges{
		Edges: edges,
	}
}

func (n *Graph) GetEdge(path *apipb.Path) (*apipb.Edge, bool) {
	if c, ok := n.edges[path.GetType()]; ok {
		node := c[path.GetId()]
		return c[path.GetId()], node != nil
	}
	return nil, false
}

func (n *Graph) SetEdge(value *apipb.Edge) *apipb.Edge {
	if value.GetId() == "" {
		value.Id = UUID()
	}
	if value.GetType() == "" {
		value.Type = apipb.Keyword_DEFAULT.String()
	}
	if _, ok := n.edges[value.GetType()]; !ok {
		n.edges[value.GetType()] = map[string]*apipb.Edge{}
	}

	n.edges[value.GetType()][value.GetId()] = value

	n.edgesFrom[value.GetFrom()] = append(n.edgesFrom[value.GetFrom()], value.Path())
	n.edgesTo[value.GetTo()] = append(n.edgesTo[value.GetTo()], value.Path())

	return value
}

func (n *Graph) RangeEdges(edgeType string, f func(edge *apipb.Edge) bool) {
	if edgeType == apipb.Keyword_ANY.String() {
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

func (n *Graph) DeleteEdge(path *apipb.Path) *apipb.Counter {
	edge, ok := n.GetEdge(path)
	if !ok {
		return &apipb.Counter{
			Count: 0,
		}
	}
	n.edgesFrom[edge.From] = removeEdge(path, n.edgesFrom[edge.From])
	n.edgesTo[edge.From] = removeEdge(path, n.edgesTo[edge.From])
	n.edgesFrom[edge.To] = removeEdge(path, n.edgesFrom[edge.To])
	n.edgesTo[edge.To] = removeEdge(path, n.edgesTo[edge.To])
	delete(n.edges[path.Type], path.Id)
	return &apipb.Counter{
		Count: 1,
	}
}

func (n *Graph) HasEdge(path *apipb.Path) bool {
	_, ok := n.GetEdge(path)
	return ok
}

func (e *Graph) RangeFrom(path *apipb.Path, fn func(e *apipb.Edge) bool) {
	for _, edge := range e.edgesFrom[path] {
		e, ok := e.GetEdge(edge)
		if ok {
			if !fn(e) {
				return
			}
		}
	}
}

func (e *Graph) RangeTo(path *apipb.Path, fn func(e *apipb.Edge) bool) {
	for _, edge := range e.edgesTo[path] {
		e, ok := e.GetEdge(edge)
		if ok {
			if !fn(e) {
				return
			}
		}
	}
}

func (e *Graph) RangeFilterFrom(path *apipb.Path, filter *apipb.TypeFilter) *apipb.Edges {
	var edges []*apipb.Edge
	e.RangeFrom(path, func(e *apipb.Edge) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Edges{
		Edges: edges,
	}
}

func (e *Graph) RangeFilterTo(path *apipb.Path, filter *apipb.TypeFilter) *apipb.Edges {
	var edges []*apipb.Edge
	e.RangeTo(path, func(e *apipb.Edge) bool {
		if e.GetType() != filter.Type {
			return true
		}
		pass, _ := BooleanExpression(filter.Expressions, e)
		if pass {
			edges = append(edges, e)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Edges{
		Edges: edges,
	}
}

func (n *Graph) FilterEdges(edgeType string, filter func(edge *apipb.Edge) bool) *apipb.Edges {
	var filtered []*apipb.Edge
	n.RangeEdges(edgeType, func(node *apipb.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return &apipb.Edges{
		Edges: filtered,
	}
}

func (n *Graph) SetEdges(edges []*apipb.Edge) *apipb.Edges {
	var returned []*apipb.Edge
	for _, edge := range edges {
		returned = append(returned, n.SetEdge(edge))
	}
	return &apipb.Edges{
		Edges: returned,
	}
}

func (n *Graph) DeleteEdges(edges []*apipb.Path) *apipb.Counter {
	count := int64(0)
	for _, edge := range edges {
		count += n.DeleteEdge(edge).Count
	}
	return &apipb.Counter{
		Count: count,
	}
}

func (n *Graph) ClearEdges(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.DeleteEdge(v.Path())
		}
	}
}

func (e *Graph) PatchEdge(value *apipb.Edge) *apipb.Edge {
	if _, ok := e.edges[value.GetType()]; !ok {
		return nil
	}
	edge, ok := e.GetEdge(value.Path())
	if !ok {
		return nil
	}
	for k, v := range value.GetAttributes().GetFields() {
		edge.GetAttributes().GetFields()[k] = v
	}
	e.SetEdge(edge)
	return edge
}

func (e *Graph) PatchEdges(values []*apipb.Edge) *apipb.Edges {
	var edges []*apipb.Edge
	for _, val := range values {
		edges = append(edges, e.PatchEdge(val))
	}
	return &apipb.Edges{
		Edges: edges,
	}
}

func (e *Graph) FilterSearchEdges(filter *apipb.TypeFilter) *apipb.Edges {
	var edges []*apipb.Edge
	var pass bool
	e.RangeEdges(filter.Type, func(edge *apipb.Edge) bool {
		pass, _ = BooleanExpression(filter.Expressions, edge)
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	})
	return &apipb.Edges{
		Edges: edges,
	}
}

func removeEdge(path *apipb.Path, paths []*apipb.Path) []*apipb.Path {
	var newPaths []*apipb.Path
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	return newPaths
}
