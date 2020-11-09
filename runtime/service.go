package runtime

import (
	"fmt"
	"github.com/autom8ter/graphik/graph"
	"time"
)

func (f *Runtime) Node(input string) (graph.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.graph.Nodes().Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", input)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *graph.Filter) (graph.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Nodes().FilterSearch(input)
}

func (f *Runtime) Edge(input string) (graph.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.graph.Edges().Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s does not exist", input)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *graph.Filter) (graph.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().FilterSearch(input)
}

func (f *Runtime) EdgesFrom(path string, filter *graph.Filter) (graph.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path string, filter *graph.Filter) (graph.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNodes(nodes graph.ValueSet) (graph.ValueSet, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_CREATE_NODES,
		Val:       nodes,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.(graph.ValueSet), nil
}

func (r *Runtime) PatchNodes(patches graph.ValueSet) (graph.ValueSet, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_PATCH_NODES,
		Val:       patches,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(graph.ValueSet), nil
}

func (r *Runtime) DelNodes(paths []string) (int, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_DELETE_NODES,
		Val:       paths,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return 0, err
	}
	if err := resp.(error); err != nil {
		return 0, err
	}
	return resp.(int), nil
}

func (r *Runtime) CreateEdges(edges graph.ValueSet) (graph.ValueSet, error) {
	for _, edge := range edges {
		if edge.GetType() == "" {
			edge.SetType(graph.Default)
		}
		if edge.GetID() == "" {
			edge.SetID(graph.UUID())
		}
		edge.SetCreatedAt(time.Now())
		edge.SetUpdatedAt(time.Now())
	}
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_CREATE_EDGES,
		Val:       edges,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(graph.ValueSet), nil
}

func (r *Runtime) PatchEdges(patch graph.ValueSet) (graph.ValueSet, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_PATCH_EDGES,
		Val:       patch,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(graph.ValueSet), nil
}

func (r *Runtime) DelEdges(paths []string) (int, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_DELETE_EDGES,
		Val:       paths,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return 0, err
	}
	if err := resp.(error); err != nil {
		return 0, err
	}
	return resp.(int), nil
}
