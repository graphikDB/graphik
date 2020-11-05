package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

func (f *Runtime) Node(input *apipb.Path) (*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.nodes.Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s.%s does not exist", input.Type, input.ID)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *apipb.Filter) ([]*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.nodes.FilterSearch(input)
}

func (f *Runtime) Edge(input *apipb.Path) (*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.edges.Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s.%s does not exist", input.Type, input.ID)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.FilterSearch(input)
}

func (f *Runtime) EdgesFrom(path *apipb.Path, filter *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path *apipb.Path, filter *apipb.Filter) ([]*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.edges.RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNode(node *apipb.Node) (*apipb.Node, error) {
	if node.Path.ID == "" {
		node.Path.ID = apipb.UUID()
	}
	if node.Path.Type == "" {
		node.Path.Type = apipb.Keyword_DEFAULT.String()
	}
	node.CreatedAt = &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	node.UpdatedAt = &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	any, err := ptypes.MarshalAny(node)
	if err != nil {

	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_CREATE_NODE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Node), nil
}

func (r *Runtime) PatchNode(patch *apipb.Patch) (*apipb.Node, error) {
	any, err := ptypes.MarshalAny(patch)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_PATCH_NODE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Node), nil
}

func (r *Runtime) DelNode(path *apipb.Path) (*apipb.Counter, error) {
	any, err := ptypes.MarshalAny(path)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_DELETE_NODE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Counter), nil
}

func (r *Runtime) CreateEdge(edge *apipb.Edge) (*apipb.Edge, error) {
	if edge.Path.ID == "" {
		edge.Path.ID = apipb.UUID()
	}
	if edge.Path.Type == "" {
		edge.Path.Type = apipb.Keyword_DEFAULT.String()
	}
	edge.CreatedAt = &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	edge.UpdatedAt = &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	any, err := ptypes.MarshalAny(edge)
	if err != nil {

	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_CREATE_EDGE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Edge), nil
}

func (r *Runtime) PatchEdge(patch *apipb.Patch) (*apipb.Edge, error) {
	any, err := ptypes.MarshalAny(patch)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_PATCH_EDGE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Edge), nil
}

func (r *Runtime) DelEdge(path *apipb.Path) (*apipb.Counter, error) {
	any, err := ptypes.MarshalAny(path)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_DELETE_EDGE,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err := resp.(error); err != nil {
		return nil, err
	}
	return resp.(*apipb.Counter), nil
}
