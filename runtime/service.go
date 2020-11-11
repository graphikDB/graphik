package runtime

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
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

func (f *Runtime) Nodes(input *apipb.TypeFilter) (*apipb.Nodes, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	nodes := f.graph.FilterSearchNodes(input)
	nodes.Sort()
	return nodes, nil
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
	edges := f.graph.FilterSearchEdges(input)
	edges.Sort()
	return edges
}

func (f *Runtime) EdgesFrom(filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges := f.graph.RangeFilterFrom(filter)
	edges.Sort()
	return edges, nil
}

func (f *Runtime) EdgesTo(filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges := f.graph.RangeFilterTo(filter)
	edges.Sort()
	return edges, nil
}

func (r *Runtime) CreateNodes(nodes *apipb.Nodes) (*apipb.Nodes, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_CREATE_NODES,
		Log: &apipb.Log{
			Log: &apipb.Log_Nodes{Nodes: nodes},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	respNodes := resp.GetNodes()
	respNodes.Sort()
	return respNodes, nil
}

func (r *Runtime) CreateNode(node *apipb.Node) (*apipb.Node, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_CREATE_NODE,
		Log: &apipb.Log{
			Log: &apipb.Log_Node{Node: node},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNode(), nil
}

func (r *Runtime) PatchNodes(nodes *apipb.Nodes) (*apipb.Nodes, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_PATCH_NODES,
		Log: &apipb.Log{
			Log: &apipb.Log_Nodes{Nodes: nodes},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	respNodes := resp.GetNodes()
	respNodes.Sort()
	return respNodes, nil
}

func (r *Runtime) PatchNode(patch *apipb.Node) (*apipb.Node, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_PATCH_NODE,
		Log: &apipb.Log{
			Log: &apipb.Log_Node{Node: patch},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNode(), nil
}

func (r *Runtime) DelNodes(paths *apipb.Paths) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Log: &apipb.Log{
			Log: &apipb.Log_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}

func (r *Runtime) DelNode(path *apipb.Path) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Log: &apipb.Log{
			Log: &apipb.Log_Path{Path: path},
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
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_CREATE_EDGES,
		Log: &apipb.Log{
			Log: &apipb.Log_Edges{Edges: edges},
		},
		Timestamp: now,
	})
	if err != nil {
		return nil, err
	}
	redges := resp.GetEdges()
	redges.Sort()
	return redges, nil
}

func (r *Runtime) CreateEdge(edge *apipb.Edge) (*apipb.Edge, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_CREATE_EDGE,
		Log: &apipb.Log{
			Log: &apipb.Log_Edge{Edge: edge},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdge(), nil
}

func (r *Runtime) PatchEdges(patches *apipb.Edges) (*apipb.Edges, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_PATCH_EDGES,
		Log: &apipb.Log{
			Log: &apipb.Log_Edges{Edges: patches},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	redges := resp.GetEdges()
	redges.Sort()
	return redges, nil
}

func (r *Runtime) PatchEdge(patch *apipb.Edge) (*apipb.Edge, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_PATCH_EDGE,
		Log: &apipb.Log{
			Log: &apipb.Log_Edge{Edge: patch},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetEdge(), nil
}

func (r *Runtime) DelEdges(paths *apipb.Paths) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Log: &apipb.Log{
			Log: &apipb.Log_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}

func (r *Runtime) DelEdge(path *apipb.Path) (*apipb.Counter, error) {
	resp, err := r.execute(&apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Log: &apipb.Log{
			Log: &apipb.Log_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCounter(), nil
}
