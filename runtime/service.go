package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/values"
	"github.com/golang/protobuf/ptypes"
	"time"
)

func (f *Runtime) Node(input string) (values.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.graph.GetNode(input)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", input)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *apipb.TypeFilter) (values.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.FilterSearchNodes(input)
}

func (f *Runtime) Edge(input string) (values.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.graph.GetEdge(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s does not exist", input)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *graph.Filter) (values.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.FilterSearchEdges(input)
}

func (f *Runtime) EdgesFrom(path string, filter *graph.Filter) (values.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path string, filter *graph.Filter) (values.ValueSet, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNodes(nodes values.ValueSet) (values.ValueSet, error) {
	any, err := ptypes.MarshalAny(&apipb.ValueSet{
		Values: nodes.Structs(),
	})
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:        apipb.Op_CREATE_NODES,
		Val:       any,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.(values.ValueSet), nil
}

func (r *Runtime) CreateNode(node values.Values) (values.Values, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_CREATE_NODES,
		Val:       values.ValueSet{node},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.(values.ValueSet)[0], nil
}

func (r *Runtime) PatchNodes(patches values.ValueSet) (values.ValueSet, error) {
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
	return resp.(values.ValueSet), nil
}

func (r *Runtime) PatchNode(patch values.Values) (values.Values, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_PATCH_NODES,
		Val:       values.ValueSet{patch},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.(values.ValueSet)[0], nil
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

func (r *Runtime) DelNode(path string) (int, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_DELETE_NODES,
		Val:       []string{path},
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

func (r *Runtime) CreateEdges(edges values.ValueSet) (values.ValueSet, error) {
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
	return resp.(values.ValueSet), nil
}

func (r *Runtime) CreateEdge(edge values.Values) (values.Values, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_CREATE_EDGES,
		Val:       values.ValueSet{edge},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(values.ValueSet)[0], nil
}

func (r *Runtime) PatchEdges(patch values.ValueSet) (values.ValueSet, error) {
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
	return resp.(values.ValueSet), nil
}

func (r *Runtime) PatchEdge(patch values.Values) (values.Values, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_PATCH_EDGES,
		Val:       values.ValueSet{patch},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(values.ValueSet)[0], nil
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

func (r *Runtime) DelEdge(path string) (int, error) {
	resp, err := r.execute(&graph.Command{
		Op:        graph.Op_DELETE_EDGES,
		Val:       []string{path},
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
