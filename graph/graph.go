package graph

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/storage"
	"github.com/autom8ter/graphik/vm"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type Graph struct {
	db        *storage.GraphStore
	edgesTo   map[string][]*apipb.Path
	edgesFrom map[string][]*apipb.Path
}

func New(path string) (*Graph, error) {
	db, err := storage.NewGraphStore(path)
	if err != nil {
		return nil, err
	}
	return &Graph{
		db:        db,
		edgesTo:   map[string][]*apipb.Path{},
		edgesFrom: map[string][]*apipb.Path{},
	}, nil
}

func (g *Graph) Do(fn func(g *Graph)) {
	fn(g)
}

//
//func (n *Graph) NodeTypes() []string {
//	var nodeTypes []string
//	for k, _ := range n.nodes {
//		nodeTypes = append(nodeTypes, k)
//	}
//	sort.Strings(nodeTypes)
//	return nodeTypes
//}

func (n *Graph) AllNodes(ctx context.Context) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	if err := n.RangeNode(ctx, apipb.Any, func(node *apipb.Node) bool {
		nodes = append(nodes, node)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Nodes{
		Nodes: nodes,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	if n.HasNode(ctx, path) {
		return n.db.GetNode(ctx, path)
	}
	return nil, noExist(path)
}

func (n *Graph) SetNode(ctx context.Context, value *apipb.Node) (*apipb.Node, error) {
	now := time.Now().UnixNano()
	if value.GetPath() == nil {
		value.Path = &apipb.Path{}
	}
	if value.GetPath().GetGid() == "" {
		value.Path.Gid = uuid.New().String()
	}
	if value.GetCreatedAt() == 0 {
		value.CreatedAt = now
	}
	if value.GetUpdatedAt() == 0 {
		value.UpdatedAt = now
	}
	if err := n.db.SetNode(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (n *Graph) PatchNode(ctx context.Context, value *apipb.Patch) (*apipb.Node, error) {
	node, err := n.db.GetNode(ctx, value.GetPath())
	if err != nil {
		return nil, err
	}
	for k, v := range value.GetAttributes().GetFields() {
		node.GetAttributes().GetFields()[k] = v
	}
	node.UpdatedAt = time.Now().UnixNano()
	return node, n.db.SetNode(ctx, node)
}

func (n *Graph) PatchNodes(ctx context.Context, values *apipb.Patches) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	for _, val := range values.GetPatches() {
		patch, err := n.PatchNode(ctx, val)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, patch)
	}
	toReturn := &apipb.Nodes{
		Nodes: nodes,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) RangeNode(ctx context.Context, nodeType string, f func(node *apipb.Node) bool) error {
	return n.db.RangeNodes(ctx, nodeType, f)
}

func (n *Graph) DeleteNode(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	if !n.HasNode(ctx, path) {
		return &empty.Empty{}, noExist(path)
	}
	var err error
	if err := n.RangeFrom(ctx, path, func(e *apipb.Edge) bool {
		_, err = n.DeleteEdge(ctx, e.GetPath())
		if e.Cascade == apipb.Cascade_CASCADE_TO || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			if _, err = n.DeleteNode(ctx, e.To); err != nil {
				err = errors.Wrap(err, err.Error())
			}
		}
		return true
	}); err != nil {
		return nil, err
	}
	if err := n.RangeTo(ctx, path, func(e *apipb.Edge) bool {
		_, err = n.DeleteEdge(ctx, e.GetPath())
		if e.Cascade == apipb.Cascade_CASCADE_FROM || e.Cascade == apipb.Cascade_CASCADE_MUTUAL {
			if _, err = n.DeleteNode(ctx, e.From); err != nil {
				err = errors.Wrap(err, err.Error())
			}
		}
		return true
	}); err != nil {
		return nil, err
	}
	if _, err = n.DeleteNode(ctx, path); err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, err
}

func (n *Graph) HasNode(ctx context.Context, path *apipb.Path) bool {
	node, _ := n.db.GetNode(ctx, path)
	return node != nil
}

func (n *Graph) FilterNode(ctx context.Context, nodeType string, filter func(node *apipb.Node) bool) (*apipb.Nodes, error) {
	var filtered []*apipb.Node
	if err := n.RangeNode(ctx, nodeType, func(node *apipb.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	}); err != nil {
		return nil, err
	}
	toreturn := &apipb.Nodes{
		Nodes: filtered,
	}
	toreturn.Sort()
	return toreturn, nil
}

func (n *Graph) SetNodes(nodes []*apipb.Node) (*apipb.Nodes, error) {
	toReturn, err := n.SetNodes(nodes)
	if err != nil {
		return nil, err
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) DeleteNodes(ctx context.Context, nodes []*apipb.Path) (*empty.Empty, error) {
	return &empty.Empty{}, n.db.DelNodes(ctx, nodes...)
}

func (n *Graph) ClearNodes(ctx context.Context, nodeType string) error {
	return n.db.DelNodeType(ctx, nodeType)
}

func (n *Graph) FilterSearchNodes(ctx context.Context, filter *apipb.Filter) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	if err := n.RangeNode(ctx, filter.Gtype, func(node *apipb.Node) bool {
		pass, err := vm.Eval(programs, node)
		if err != nil {
			return true
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Nodes{
		Nodes: nodes,
	}
	toReturn.Sort()
	return toReturn, nil
}

//func (n *Graph) EdgeTypes() []string {
//	var edgeTypes []string
//	for k, _ := range n.edges {
//		edgeTypes = append(edgeTypes, k)
//	}
//	sort.Strings(edgeTypes)
//	return edgeTypes
//}

func (n *Graph) AllEdges(ctx context.Context) (*apipb.Edges, error) {
	var edges []*apipb.Edge
	if err := n.RangeEdges(ctx, apipb.Any, func(edge *apipb.Edge) bool {
		edges = append(edges, edge)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	return n.db.GetEdge(ctx, path)
}

func (n *Graph) SetEdge(ctx context.Context, value *apipb.Edge) (*apipb.Edge, error) {
	if value.Path == nil {
		value.Path = &apipb.Path{}
	}
	if value.GetPath().GetGid() == "" {
		value.Path.Gid = uuid.New().String()
	}
	if err := n.db.SetEdge(ctx, value); err != nil {
		return nil, err
	}
	n.edgesFrom[value.GetFrom().String()] = append(n.edgesFrom[value.GetFrom().String()], value.GetPath())
	n.edgesTo[value.GetTo().String()] = append(n.edgesTo[value.GetTo().String()], value.GetPath())

	return value, nil
}

func (n *Graph) RangeEdges(ctx context.Context, edgeType string, f func(edge *apipb.Edge) bool) error {
	return n.db.RangeEdges(ctx, edgeType, f)
}

func (n *Graph) DeleteEdge(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	edge, err := n.GetEdge(ctx, path)
	if err != nil {
		return nil, err
	}
	n.edgesFrom[edge.From.String()] = removeEdge(path, n.edgesFrom[edge.From.String()])
	n.edgesTo[edge.From.String()] = removeEdge(path, n.edgesTo[edge.From.String()])
	n.edgesFrom[edge.To.String()] = removeEdge(path, n.edgesFrom[edge.To.String()])
	n.edgesTo[edge.To.String()] = removeEdge(path, n.edgesTo[edge.To.String()])
	if err := n.db.DelEdges(ctx, path); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (n *Graph) HasEdge(ctx context.Context, path *apipb.Path) bool {
	e, _ := n.db.GetEdge(ctx, path)
	return e != nil
}

func (g *Graph) RangeFrom(ctx context.Context, path *apipb.Path, fn func(e *apipb.Edge) bool) error {
	for _, path := range g.edgesFrom[path.String()] {
		edge, err := g.GetEdge(ctx, path)
		if err != nil {
			return err
		}
		if !fn(edge) {
			return nil
		}
	}
	return nil
}

func (g *Graph) RangeTo(ctx context.Context, path *apipb.Path, fn func(edge *apipb.Edge) bool) error {
	for _, edge := range g.edgesTo[path.String()] {
		e, err := g.GetEdge(ctx, edge)
		if err != nil {
			return err
		}
		if !fn(e) {
			return nil
		}
	}
	return nil
}

func (g *Graph) RangeFilterFrom(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err = g.RangeFrom(ctx, filter.NodePath, func(edge *apipb.Edge) bool {
		if filter.Gtype != "*" {
			if edge.GetPath().GetGtype() != filter.Gtype {
				return true
			}
		}

		pass, err = vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, err
}

func (e *Graph) RangeFilterTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err := e.RangeTo(ctx, filter.NodePath, func(edge *apipb.Edge) bool {
		if filter.Gtype != "*" {
			if edge.GetPath().GetGtype() != filter.Gtype {
				return true
			}
		}
		pass, err = vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) FilterEdges(ctx context.Context, edgeType string, filter func(edge *apipb.Edge) bool) (*apipb.Edges, error) {
	var filtered []*apipb.Edge
	if err := n.RangeEdges(ctx, edgeType, func(node *apipb.Edge) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: filtered,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) SetEdges(ctx context.Context, edges []*apipb.Edge) (*apipb.Edges, error) {
	var returned []*apipb.Edge
	for _, edge := range edges {
		e, err := n.SetEdge(ctx, edge)
		if err != nil {
			return nil, err
		}
		returned = append(returned, e)
	}
	toReturn := &apipb.Edges{
		Edges: returned,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *Graph) DeleteEdges(ctx context.Context, edges []*apipb.Path) (*empty.Empty, error) {
	for _, edge := range edges {
		_, err := n.DeleteEdge(ctx, edge)
		if err != nil {
			return nil, err
		}
	}
	return &empty.Empty{}, nil
}

func (n *Graph) ClearEdges(ctx context.Context, edgeType string) error {
	return n.db.DelNodeType(ctx, edgeType)
}

func (e *Graph) PatchEdge(ctx context.Context, value *apipb.Patch) (*apipb.Edge, error) {
	edge, err := e.GetEdge(ctx, value.GetPath())
	if err != nil {
		return nil, err
	}
	for k, v := range value.GetAttributes().GetFields() {
		edge.GetAttributes().GetFields()[k] = v
	}
	edge, err = e.SetEdge(ctx, edge)
	if err != nil {
		return nil, err
	}
	return edge, nil
}

func (e *Graph) PatchEdges(ctx context.Context, values *apipb.Patches) (*apipb.Edges, error) {
	var edges []*apipb.Edge
	for _, val := range values.GetPatches() {
		patch, err := e.PatchEdge(ctx, val)
		if err != nil {
			return nil, err
		}
		edges = append(edges, patch)
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (e *Graph) FilterSearchEdges(ctx context.Context, filter *apipb.Filter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	if err := e.RangeEdges(ctx, filter.Gtype, func(edge *apipb.Edge) bool {
		pass, err := vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (g *Graph) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	graph := &apipb.Graph{
		Nodes: &apipb.Nodes{},
		Edges: &apipb.Edges{},
	}
	nodes, err := g.FilterSearchNodes(ctx, filter.Nodes)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.GetNodes() {
		graph.Nodes.Nodes = append(graph.Nodes.Nodes, node)
		edges, err := g.RangeFilterFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdges().GetGtype(),
			Expressions: filter.GetEdges().GetExpressions(),
			Limit:       filter.GetEdges().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		graph.Edges.Edges = append(graph.Edges.Edges, edges.GetEdges()...)
	}
	graph.Edges.Sort()
	graph.Nodes.Sort()
	return graph, err
}

func (g *Graph) GetEdgeDetail(ctx context.Context, path *apipb.Path) (*apipb.EdgeDetail, error) {
	e, err := g.db.GetEdge(ctx, path)
	if err != nil {
		return nil, err
	}
	from, err := g.db.GetNode(ctx, e.From)
	if err != nil {
		return nil, err
	}
	to, err := g.db.GetNode(ctx, e.To)
	if err != nil {
		return nil, err
	}
	return &apipb.EdgeDetail{
		Path:       e.Path,
		Attributes: e.Attributes,
		Cascade:    e.Cascade,
		From:       from,
		To:         to,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}, nil
}

func (g *Graph) GetNodeDetail(ctx context.Context, filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	detail := &apipb.NodeDetail{
		Path:      filter.GetPath(),
		EdgesTo:   map[string]*apipb.EdgeDetails{},
		EdgesFrom: map[string]*apipb.EdgeDetails{},
	}
	node, err := g.db.GetNode(ctx, filter.GetPath())
	if err != nil {
		return nil, err
	}
	detail.UpdatedAt = node.UpdatedAt
	detail.CreatedAt = node.CreatedAt
	detail.Attributes = node.Attributes
	if filter.GetEdgesFrom() != nil {
		edgesFrom, err := g.RangeFilterFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdgesFrom().GetGtype(),
			Expressions: filter.GetEdgesFrom().GetExpressions(),
			Limit:       filter.GetEdgesFrom().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edgesFrom.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesFrom[edge.GetPath().GetGtype()].Edges = append(detail.EdgesFrom[edge.GetPath().GetGtype()].Edges, eDetail)
		}
	}

	if filter.GetEdgesTo() != nil {
		edgesTo, err := g.RangeFilterTo(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdgesTo().GetGtype(),
			Expressions: filter.GetEdgesTo().GetExpressions(),
			Limit:       filter.GetEdgesTo().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edgesTo.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesTo[edge.GetPath().GetGtype()].Edges = append(detail.EdgesTo[edge.GetPath().GetGtype()].Edges, eDetail)
		}
	}
	for _, d := range detail.EdgesTo {
		d.Sort()
	}
	for _, d := range detail.EdgesFrom {
		d.Sort()
	}
	return detail, nil
}

func (g *Graph) Close() error {
	return g.db.Close()
}

func (g *Graph) Schema(ctx context.Context) (*apipb.Schema, error) {
	etypes, err := g.db.EdgeTypes(ctx)
	if err != nil {
		return nil, err
	}
	ntypes, err := g.db.NodeTypes(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Schema{
		EdgeTypes: etypes,
		NodeTypes: ntypes,
	}, nil
}
