package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/graph"
	"time"
)

func (f *Runtime) Node(input *apipb.Path) (*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.graph.GetNode(input)
	if !ok {
		return nil, fmt.Errorf("node %s does not exist", input)
	}
	return node, nil
}

func (f *Runtime) Nodes(input *apipb.TypeFilter) ([]*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.FilterSearchNodes(input)
}

func (f *Runtime) Edge(input *apipb.Path) (*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.graph.GetEdge(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s does not exist", input)
	}
	return edge, nil
}

func (f *Runtime) Edges(input *apipb.TypeFilter) *apipb.Edges {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.FilterSearchEdges(input)
}

func (f *Runtime) EdgesFrom(path *apipb.Path, filter *apipb.TypeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.RangeFilterFrom(path, filter), nil
}

func (f *Runtime) EdgesTo(path *apipb.Path, filter *apipb.TypeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.RangeFilterTo(path, filter), nil
}

func (r *Runtime) CreateNodes(nodes *apipb.Nodes) (*apipb.Nodes, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_CREATE_NODES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Nodes{Nodes: nodes},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNodes(), nil
}

func (r *Runtime) CreateNode(node *apipb.Node) (*apipb.Node, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_CREATE_NODE,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Node{Node: node},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNode(), nil
}

func (r *Runtime) PatchNodes(nodes *apipb.Nodes) (*apipb.Nodes, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_PATCH_NODES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Nodes{Nodes: nodes},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNodes(), nil
}

func (r *Runtime) PatchNode(patch *apipb.Node) (*apipb.Node, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_PATCH_NODE,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Node{Node: patch},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNode(), nil
}

func (r *Runtime) DelNodes(paths *apipb.Paths) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_DELETE_NODES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}

func (r *Runtime) DelNode(path *apipb.Path) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_DELETE_NODES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}

func (r *Runtime) CreateEdges(edges *apipb.Edges) (*apipb.Edges, error) {
	now := time.Now().UnixNano()
	for _, edge := range edges.GetEdges() {
		if edge.GetType() == "" {
			edge.Type = apipb.Keyword_DEFAULT.String()
		}
		if edge.GetId() == "" {
			edge.Id = graph.UUID()
		}
		edge.CreatedAt = now
		edge.UpdatedAt = now
	}
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_CREATE_EDGES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Edges{Edges: edges},
		},
		Timestamp: now,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdges(), nil
}

func (r *Runtime) CreateEdge(edge *apipb.Edge) (*apipb.Edge, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_CREATE_EDGE,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Edge{Edge: edge},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdge(), nil
}

func (r *Runtime) PatchEdges(patches *apipb.Edges) (*apipb.Edges, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_PATCH_EDGES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Edges{Edges: patches},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdges(), nil
}

func (r *Runtime) PatchEdge(patch *apipb.Edge) (*apipb.Edge, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_PATCH_EDGE,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Edge{Edge: patch},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdge(), nil
}

func (r *Runtime) DelEdges(paths *apipb.Paths) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_DELETE_EDGES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}

func (r *Runtime) DelEdge(path *apipb.Path) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.Command{
		Op: apipb.Op_DELETE_EDGES,
		Val: &apipb.RaftLog{
			Log: &apipb.RaftLog_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}
