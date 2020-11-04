package runtime

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/model"
)

func (f *Runtime) Node(ctx context.Context, input model.Path) (*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.nodes.Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s.%s does not exist", input.Type, input.ID)
	}
	return node, nil
}

func (f *Runtime) Nodes(ctx context.Context, input model.Filter) ([]*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.nodes.FilterSearch(input)
}

func (f *Runtime) DepthTo(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var nodes []*model.Node
	if err := f.nodes.RangeToDepth(input, func(node *model.Node) bool {
		nodes = append(nodes, node)
		return len(nodes) < input.Limit
	}); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (f *Runtime) DepthFrom(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var nodes []*model.Node
	if err := f.nodes.RangeFromDepth(input, func(node *model.Node) bool {
		nodes = append(nodes, node)
		return len(nodes) < input.Limit
	}); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (f *Runtime) Edge(ctx context.Context, input model.Path) (*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.edges.Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s.%s does not exist", input.Type, input.ID)
	}
	return edge, nil
}

func (f *Runtime) Edges(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.FilterSearch(input)
}
