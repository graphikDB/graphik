package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/express"
	"github.com/google/uuid"
	"sort"
	"time"
)

type Graph struct {
	nodes     map[string]map[string]*apipb.Node
	edges     map[string]map[string]*apipb.Edge
	edgesTo   map[string][]*apipb.Path
	edgesFrom map[string][]*apipb.Path
}

func New() *Graph {
	return &Graph{
		nodes:     map[string]map[string]*apipb.Node{},
		edges:     map[string]map[string]*apipb.Edge{},
		edgesTo:   map[string][]*apipb.Path{},
		edgesFrom: map[string][]*apipb.Path{},
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
	if c, ok := n.nodes[path.GetGtype()]; ok {
		node := c[path.GetGid()]
		return node, node != nil
	}
	return nil, false
}

func (n *Graph) SetNode(value *apipb.Node) *apipb.Node {
	now := time.Now().UnixNano()
	if value.GetPath() == nil {
		value.Path = &apipb.Path{}
	}
	if value.GetPath().GetGid() == "" {
		value.Path.Gid = uuid.New().String()
	}
	if value.GetPath().GetGtype() == "" {
		value.Path.Gtype = apipb.Keyword_DEFAULT.String()
	}
	if value.GetCreatedAt() == 0 {
		value.CreatedAt = now
	}
	if value.GetUpdatedAt() == 0 {
		value.UpdatedAt = now
	}
	if nodeMap, ok := n.nodes[value.GetPath().GetGtype()]; !ok {
		n.nodes[value.GetPath().GetGtype()] = map[string]*apipb.Node{}
	} else {
		nodeMap[value.GetPath().GetGid()] = value
	}
	return value
}

func (n *Graph) PatchNode(value *apipb.Node) *apipb.Node {
	value.UpdatedAt = time.Now().UnixNano()
	if _, ok := n.nodes[value.GetPath().GetGtype()]; !ok {
		return nil
	}
	node := n.nodes[value.GetPath().GetGtype()][value.GetPath().GetGid()]
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
	n.RangeFrom(path, 0, func(e *apipb.Edge) bool {
		n.DeleteEdge(e.GetPath())
		if e.Cascade == apipb.Cascade_CASCADE_TO || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			n.DeleteNode(e.To)
		}
		return true
	})
	n.RangeTo(path, 0, func(e *apipb.Edge) bool {
		n.DeleteEdge(e.GetPath())
		if e.Cascade == apipb.Cascade_CASCADE_FROM || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			n.DeleteNode(e.From)
		}
		return true
	})

	delete(n.nodes[path.Gtype], path.Gid)
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

func (n *Graph) FilterSearchNodes(filter *apipb.TypeFilter) *apipb.Nodes {
	var nodes []*apipb.Node
	n.RangeNode(filter.Gtype, func(node *apipb.Node) bool {
		pass, err := express.Eval(filter.Expressions, node)
		if err != nil {
			panic(err)
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})

	return &apipb.Nodes{
		Nodes: nodes,
	}
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
	if c, ok := n.edges[path.GetGtype()]; ok {
		node := c[path.GetGid()]
		return c[path.GetGid()], node != nil
	}
	return nil, false
}

func (n *Graph) SetEdge(value *apipb.Edge) *apipb.Edge {
	if value.Path == nil {
		value.Path = &apipb.Path{}
	}
	if value.GetPath().GetGid() == "" {
		value.Path.Gid = uuid.New().String()
	}
	if value.GetPath().GetGtype() == "" {
		value.Path.Gtype = apipb.Keyword_DEFAULT.String()
	}
	if _, ok := n.edges[value.GetPath().GetGtype()]; !ok {
		n.edges[value.GetPath().GetGtype()] = map[string]*apipb.Edge{}
	}

	n.edges[value.GetPath().GetGtype()][value.GetPath().GetGid()] = value

	n.edgesFrom[value.GetFrom().String()] = append(n.edgesFrom[value.GetFrom().String()], value.GetPath())
	n.edgesTo[value.GetTo().String()] = append(n.edgesTo[value.GetTo().String()], value.GetPath())

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
	n.edgesFrom[edge.From.String()] = removeEdge(path, n.edgesFrom[edge.From.String()])
	n.edgesTo[edge.From.String()] = removeEdge(path, n.edgesTo[edge.From.String()])
	n.edgesFrom[edge.To.String()] = removeEdge(path, n.edgesFrom[edge.To.String()])
	n.edgesTo[edge.To.String()] = removeEdge(path, n.edgesTo[edge.To.String()])
	delete(n.edges[path.Gtype], path.Gid)
	return &apipb.Counter{
		Count: 1,
	}
}

func (n *Graph) HasEdge(path *apipb.Path) bool {
	_, ok := n.GetEdge(path)
	return ok
}

func (g *Graph) RangeFrom(path *apipb.Path, degree int32, fn func(e *apipb.Edge) bool) {
	visited := map[*apipb.Path]struct{}{}
	for _, path := range g.edgesFrom[path.String()] {
		edge, _ := g.GetEdge(path)
		for x := int32(0); x < degree; x++ {
			fn = EdgeFunc(fn).ascendFrom(g, visited)
		}
		if !fn(edge) {
			return
		}
	}
}

func (g *Graph) RangeTo(path *apipb.Path, degree int32, fn func(edge *apipb.Edge) bool) {
	visited := map[*apipb.Path]struct{}{}
	for _, edge := range g.edgesTo[path.String()] {
		e, ok := g.GetEdge(edge)
		if ok {
			for x := int32(0); x < degree; x++ {
				fn = EdgeFunc(fn).ascendTo(g, visited)
			}
			if !fn(e) {
				return
			}
		}
	}
}

func (g *Graph) RangeFilterFrom(filter *apipb.EdgeFilter) *apipb.Edges {
	var edges []*apipb.Edge

	g.RangeFrom(filter.NodePath, filter.MaxDegree, func(edge *apipb.Edge) bool {
		if edge.GetPath().GetGtype() != filter.Gtype {
			return true
		}
		pass, err := express.Eval(filter.Expressions, edge)
		if err != nil {
			panic(err)
		}
		if pass {
			edges = append(edges, edge)
		}

		return len(edges) < int(filter.Limit)
	})
	return &apipb.Edges{
		Edges: edges,
	}
}

func (e *Graph) RangeFilterTo(filter *apipb.EdgeFilter) *apipb.Edges {
	var edges []*apipb.Edge
	e.RangeTo(filter.NodePath, filter.MaxDegree, func(edge *apipb.Edge) bool {
		if edge.GetPath().GetGtype() != filter.Gtype {
			return true
		}
		pass, err := express.Eval(filter.Expressions, edge)
		if err != nil {
			panic(err)
		}
		if pass {
			edges = append(edges, edge)
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
			n.DeleteEdge(v.GetPath())
		}
	}
}

func (e *Graph) PatchEdge(value *apipb.Edge) *apipb.Edge {
	if _, ok := e.edges[value.GetPath().GetGtype()]; !ok {
		return nil
	}
	edge, ok := e.GetEdge(value.GetPath())
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
	e.RangeEdges(filter.Gtype, func(edge *apipb.Edge) bool {
		pass, _ = express.Eval(filter.Expressions, edge)
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
