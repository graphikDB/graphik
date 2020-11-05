package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
)

func (f *Runtime) Node(ctx context.Context, input *apipb.Path) (*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.nodes.Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s.%s does not exist", input.Type, input.ID)
	}
	return node, nil
}

func (f *Runtime) Nodes(ctx context.Context, input *apipb.Filter) ([]*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.nodes.FilterSearch(input)
}

func (f *Runtime) Edge(ctx context.Context, input *apipb.Path) (*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.edges.Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s.%s does not exist", input.Type, input.ID)
	}
	return edge, nil
}

func (f *Runtime) Edges(ctx context.Context, input *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.FilterSearch(input)
}

func (f *Runtime) EdgesFrom(ctx context.Context, path *apipb.Path, filter *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(ctx context.Context, path *apipb.Path, filter *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.RangeFilterTo(path, filter), nil
}
