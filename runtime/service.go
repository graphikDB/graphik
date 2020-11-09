package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

func (f *Runtime) Node(input string) (*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.graph.Nodes().Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", input)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *apipb.Filter) ([]*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Nodes().FilterSearch(input)
}

func (f *Runtime) Edge(input string) (*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.graph.Edges().Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s does not exist", input)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *apipb.Filter) ([]*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().FilterSearch(input)
}

func (f *Runtime) EdgesFrom(path string, filter *apipb.Filter) ([]*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path string, filter *apipb.Filter) ([]*lang.Values, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.Edges().RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNodes(nodes *apipb.ValueSet) ([]*lang.Values, error) {
	for _, node := range nodes.Values {
		xtype, xid := lang.SplitPath(node.Path)
		if xtype == "" {
			xtype = apipb.Keyword_DEFAULT.String()
			node.Path = lang.FormPath(xtype, xid)
		}
		if xid == "" {
			xid = lang.UUID()
			node.Path = lang.FormPath(xtype, xid)
		}
		node.CreatedAt = &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		}
		node.UpdatedAt = &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		}
	}
	any, err := ptypes.MarshalAny(nodes)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_CREATE_NODES,
		Val: any,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	})
	if err != nil {
		return nil, err
	}
	if err, ok := resp.(error); ok {
		return nil, err
	}
	return resp.([]*lang.Values), nil
}

func (r *Runtime) PatchNodes(patches []*lang.Values) ([]*lang.Values, error) {
	any, err := ptypes.MarshalAny(patches)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_PATCH_NODES,
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
	return resp.([]*lang.Values), nil
}

func (r *Runtime) DelNodes(paths *apipb.Paths) (*apipb.Counter, error) {
	any, err := ptypes.MarshalAny(paths)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_DELETE_NODES,
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

func (r *Runtime) CreateEdges(edges []*lang.Values) ([]*lang.Values, error) {
	for _, edge := range edges.Edges {
		xtype, xid := lang.SplitPath(edge.Path)
		if xtype == "" {
			xtype = apipb.Keyword_DEFAULT.String()
			edge.Path = lang.FormPath(xtype, xid)
		}
		if xid == "" {
			xid = lang.UUID()
			edge.Path = lang.FormPath(xtype, xid)
		}
		edge.CreatedAt = &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		}
		edge.UpdatedAt = &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		}
	}
	any, err := ptypes.MarshalAny(edges)
	if err != nil {

	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_CREATE_EDGES,
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
	return resp.([]*lang.Values), nil
}

func (r *Runtime) PatchEdges(patch *apipb.ValueSet) ([]*lang.Values, error) {
	any, err := ptypes.MarshalAny(patch)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_PATCH_EDGES,
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
	return resp.([]*lang.Values), nil
}

func (r *Runtime) DelEdges(paths *apipb.Paths) (*apipb.Counter, error) {
	any, err := ptypes.MarshalAny(paths)
	if err != nil {
		return nil, err
	}
	resp, err := r.execute(&apipb.Command{
		Op:  apipb.Op_DELETE_EDGES,
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
