package runtime

import (
	"fmt"
	"github.com/autom8ter/graphik/lang"
	"time"
)

func (f *Runtime) Node(input string) (lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.graph.Nodes().Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", input)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *lang.Filter) (lang.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Nodes().FilterSearch(input)
}

func (f *Runtime) Edge(input string) (lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.graph.Edges().Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s does not exist", input)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *lang.Filter) (lang.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().FilterSearch(input)
}

func (f *Runtime) EdgesFrom(path string, filter *lang.Filter) (lang.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path string, filter *lang.Filter) (lang.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNodes(nodes lang.ValueSet) (lang.ValueSet, error) {
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_CREATE_NODES,
		Val:       nodes,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.(lang.ValueSet), nil
}

func (r *Runtime) PatchNodes(patches lang.ValueSet) (lang.ValueSet, error) {
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_PATCH_NODES,
		Val:       patches,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(lang.ValueSet), nil
}

func (r *Runtime) DelNodes(paths []string) (int, error) {
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_DELETE_NODES,
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

func (r *Runtime) CreateEdges(edges lang.ValueSet) (lang.ValueSet, error) {
	for _, edge := range edges {
		if edge.GetType() == "" {
			edge.SetType(lang.Default)
		}
		if edge.GetID() == "" {
			edge.SetID(lang.UUID())
		}
		edge.SetCreatedAt(time.Now())
		edge.SetUpdatedAt(time.Now())
	}
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_CREATE_EDGES,
		Val:       edges,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(lang.ValueSet), nil
}

func (r *Runtime) PatchEdges(patch lang.ValueSet) (lang.ValueSet, error) {
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_PATCH_EDGES,
		Val:       patch,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(lang.ValueSet), nil
}

func (r *Runtime) DelEdges(paths []string) (int, error) {
	resp, err := r.execute(&lang.Command{
		Op:        lang.Op_DELETE_EDGES,
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
