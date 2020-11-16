package graph

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/express"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sort"
	"time"
)

type Graph struct {
	basePath  string
	nodes     map[string]map[string]*apipb.Path
	edges     map[string]map[string]*apipb.Path
	edgesTo   map[string][]*apipb.Path
	edgesFrom map[string][]*apipb.Path
}

func New(basePath string) *Graph {
	return &Graph{
		basePath: basePath,
		nodes:     map[string]map[string]*apipb.Path{},
		edges:     map[string]map[string]*apipb.Path{},
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

func (n *Graph) GetNode(path *apipb.Path) (*apipb.Node, error) {
	if n.HasNode(path) {
		return n.getNode(path)
	}
	return nil, noExist(path)
}

func (n *Graph) SetNode(value *apipb.Node) (*apipb.Node, error) {
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
	if _, ok := n.nodes[value.GetPath().GetGtype()]; !ok {
		n.nodes[value.GetPath().GetGtype()] = map[string]*apipb.Path{}
	}
	if err := n.setNode(value); err != nil {
		return nil, err
	}
	n.nodes[value.GetPath().GetGtype()][value.GetPath().GetGid()] = value.GetPath()
	return value, nil
}

func (n *Graph) PatchNode(value *apipb.Patch) (*apipb.Node, error) {
	if _, ok := n.nodes[value.GetPath().GetGtype()]; !ok {
		return nil, noExist(value.GetPath())
	}
	node, err := n.getNode(value.GetPath())
	if err != nil {
		return nil, err
	}
	for k, v := range value.GetAttributes().GetFields() {
		node.GetAttributes().GetFields()[k] = v
	}
	node.UpdatedAt = time.Now().UnixNano()
	return node, nil
}

func (n *Graph) PatchNodes(values *apipb.Patches) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	for _, val := range values.GetPatches() {
		patch, err := n.PatchNode(val)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, patch)
	}
	return &apipb.Nodes{
		Nodes: nodes,
	}, nil
}

func (n *Graph) RangeNode(nodeType string, f func(node *apipb.Node) bool) error {
	if nodeType == apipb.Keyword_ANY.String() {
		for _, c := range n.nodes {
			for _, path := range c {
				node, err := n.getNode(path)
				if err != nil {
					return err
				}
				if !f(node) {
					return nil
				}
			}
		}
	} else {
		if c, ok := n.nodes[nodeType]; ok {
			for _, path := range c {
				node, err := n.getNode(path)
				if err != nil {
					return err
				}
				if !f(node) {
					return nil
				}
			}
		}
	}
	return nil
}

func (n *Graph) DeleteNode(path *apipb.Path) *apipb.Counter {
	if !n.HasNode(path) {
		return &apipb.Counter{
			Count: 0,
		}
	}
	n.RangeFrom(path, func(e *apipb.Edge) bool {
		n.DeleteEdge(e.GetPath())
		if e.Cascade == apipb.Cascade_CASCADE_TO || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			n.DeleteNode(e.To)
		}
		return true
	})
	n.RangeTo(path, func(e *apipb.Edge) bool {
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
	if values, ok := n.nodes[path.GetGtype()]; ok {
		if _, ok := values[path.GetGid()]; ok {
			return true
		}
	}
	return false
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

func (n *Graph) SetNodes(nodes []*apipb.Node) (*apipb.Nodes, error) {
	var returned []*apipb.Node
	for _, node := range nodes {
		node, err := n.SetNode(node)
		if err != nil {
			return nil, err
		}
		returned = append(returned, node)
	}
	return &apipb.Nodes{
		Nodes: nodes,
	}, nil
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

func (n *Graph) GetEdge(path *apipb.Path) (*apipb.Edge, error) {
	if c, ok := n.edges[path.GetGtype()]; ok {
		path := c[path.GetGid()]
		return n.getEdge(path)
	}
	return nil, noExist(path)
}

func (n *Graph) SetEdge(value *apipb.Edge) (*apipb.Edge, error) {
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
		n.edges[value.GetPath().GetGtype()] = map[string]*apipb.Path{}
	}
	if err := n.setEdge(value); err != nil {
		return nil, err
	}
	n.edges[value.GetPath().GetGtype()][value.GetPath().GetGid()] = value.GetPath()

	n.edgesFrom[value.GetFrom().String()] = append(n.edgesFrom[value.GetFrom().String()], value.GetPath())
	n.edgesTo[value.GetTo().String()] = append(n.edgesTo[value.GetTo().String()], value.GetPath())

	return value, nil
}

func (n *Graph) RangeEdges(edgeType string, f func(edge *apipb.Edge) bool) error {
	if edgeType == apipb.Keyword_ANY.String() {
		for _, c := range n.edges {
			for _, path := range c {
				e, err := n.getEdge(path)
				if err != nil {
					return err
				}
				if !f(e) {
					return nil
				}
			}
		}
	} else {
		if c, ok := n.edges[edgeType]; ok {
			for _, path := range c {
				e, err := n.getEdge(path)
				if err != nil {
					return err
				}
				if !f(e) {
					return nil
				}
			}
		}
	}
	return nil
}

func (n *Graph) DeleteEdge(path *apipb.Path) (*apipb.Counter, error) {
	edge, err := n.GetEdge(path)
	if err != nil {
		return nil, err
	}
	n.edgesFrom[edge.From.String()] = removeEdge(path, n.edgesFrom[edge.From.String()])
	n.edgesTo[edge.From.String()] = removeEdge(path, n.edgesTo[edge.From.String()])
	n.edgesFrom[edge.To.String()] = removeEdge(path, n.edgesFrom[edge.To.String()])
	n.edgesTo[edge.To.String()] = removeEdge(path, n.edgesTo[edge.To.String()])
	delete(n.edges[path.Gtype], path.Gid)
	return &apipb.Counter{
		Count: 1,
	}, nil
}

func (n *Graph) HasEdge(path *apipb.Path) bool {
	types, ok := n.edges[path.GetGtype()]
	if ok {
		_, ok := types[path.GetGid()]
		if ok {
			return true
		}
	}
	return false
}

func (g *Graph) RangeFrom(path *apipb.Path, fn func(e *apipb.Edge) bool) {
	for _, path := range g.edgesFrom[path.String()] {
		edge, _ := g.GetEdge(path)
		if !fn(edge) {
			return
		}
	}
}

func (g *Graph) RangeTo(path *apipb.Path, fn func(edge *apipb.Edge) bool) error {
	for _, edge := range g.edgesTo[path.String()] {
		e, err := g.GetEdge(edge)
		if err != nil {
			return err
		}
		if !fn(e) {
			return nil
		}
	}
	return nil
}

func (g *Graph) RangeFilterFrom(filter *apipb.EdgeFilter) *apipb.Edges {
	var edges []*apipb.Edge

	g.RangeFrom(filter.NodePath, func(edge *apipb.Edge) bool {
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
	e.RangeTo(filter.NodePath, func(edge *apipb.Edge) bool {
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

func (n *Graph) SetEdges(edges []*apipb.Edge) (*apipb.Edges, error) {
	var returned []*apipb.Edge
	for _, edge := range edges {
		e, err := n.SetEdge(edge)
		if err != nil {
			return nil, err
		}
		returned = append(returned, e)
	}
	return &apipb.Edges{
		Edges: returned,
	}, nil
}

func (n *Graph) DeleteEdges(edges []*apipb.Path) (*apipb.Counter, error) {
	count := int64(0)
	for _, edge := range edges {
		deleted, err := n.DeleteEdge(edge)
		if err != nil {
			return nil, err
		}
		count += deleted.Count
	}
	return &apipb.Counter{
		Count: count,
	}, nil
}

func (n *Graph) ClearEdges(edgeType string) {
	if cache, ok := n.edges[edgeType]; ok {
		for _, v := range cache {
			n.DeleteEdge(v)
		}
	}
}

func (e *Graph) PatchEdge(value *apipb.Patch) (*apipb.Edge, error) {
	if _, ok := e.edges[value.GetPath().GetGtype()]; !ok {
		return nil, noExist(value.GetPath())
	}
	edge, err := e.GetEdge(value.GetPath())
	if err != nil {
		return nil, err
	}
	for k, v := range value.GetAttributes().GetFields() {
		edge.GetAttributes().GetFields()[k] = v
	}
	edge, err = e.SetEdge(edge)
	if err != nil {
		return nil, err
	}
	return edge, nil
}

func (e *Graph) PatchEdges(values *apipb.Patches) (*apipb.Edges, error) {
	var edges []*apipb.Edge
	for _, val := range values.GetPatches() {
		patch, err := e.PatchEdge(val)
		if err != nil {
			return nil, err
		}
		edges = append(edges, patch)
	}
	return &apipb.Edges{
		Edges: edges,
	}, nil
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

func (g *Graph) getEdge(path *apipb.Path) (*apipb.Edge, error) {
	bits, err := ioutil.ReadFile(fmt.Sprintf("%s/edges/%s/%s", g.basePath, path.GetGtype(), path.GetGid()))
	if err != nil {
		return nil, err
	}
	var e apipb.Edge
	if err := proto.Unmarshal(bits, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (g *Graph) getNode(path *apipb.Path) (*apipb.Node, error) {
	bits, err := ioutil.ReadFile(fmt.Sprintf("%s/nodes/%s/%s", g.basePath, path.GetGtype(), path.GetGid()))
	if err != nil {
		return nil, err
	}
	var n apipb.Node
	if err := proto.Unmarshal(bits, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

func (g *Graph) setNode(node *apipb.Node) error {
	f, err := os.Create(fmt.Sprintf("%s/nodes/%s/%s", g.basePath, node.GetPath().GetGtype(), node.GetPath().GetGid()))
	defer f.Close()
	if err != nil {
		return err
	}
	bits, err := proto.Marshal(node)
	if err != nil {
		return err
	}
	_, err = f.Write(bits)
	return err
}

func (g *Graph) setEdge(node *apipb.Edge) error {
	f, err := os.Create(fmt.Sprintf("%s/edges/%s/%s", g.basePath, node.GetPath().GetGtype(), node.GetPath().GetGid()))
	defer f.Close()
	if err != nil {
		return err
	}
	bits, err := proto.Marshal(node)
	if err != nil {
		return err
	}
	_, err = f.Write(bits)
	return err
}

func noExist(path *apipb.Path) error {
	return errors.Errorf("%s.%s does not exist", path.Gtype, path.Gid)
}
