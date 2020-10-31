package store

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/graph/model"
)

func (f *Store) Node(ctx context.Context, input model.Path) (*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.nodes.Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s.%s does not exist", input.Type, input.ID)
	}
	return node, nil
}

func (f *Store) Nodes(ctx context.Context, input model.Filter) ([]*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.nodes.FilterSearch(input)
}

func (f *Store) DepthTo(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
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

func (f *Store) DepthFrom(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
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

func (f *Store) Edge(ctx context.Context, input model.Path) (*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.edges.Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s.%s does not exist", input.Type, input.ID)
	}
	return edge, nil
}

func (f *Store) Edges(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.FilterSearch(input)
}

func (f *Store) SearchNodes(ctx context.Context, input model.Search) (*model.SearchResults, error) {
	return f.nodes.Search(input.Search, input.Type)
}

func (f *Store) SearchEdges(ctx context.Context, input model.Search) (*model.SearchResults, error) {
	return f.edges.Search(input.Search, input.Type)
}
