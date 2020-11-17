package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"time"
)

func (f *Runtime) Node(input *apipb.Path) (*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.GetNode(input)
}

func (f *Runtime) Nodes(input *apipb.Filter) (*apipb.Nodes, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	nodes, err := f.graph.FilterSearchNodes(input)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return nodes, nil
}

func (f *Runtime) Edge(input *apipb.Path) (*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.GetEdge(input)
}

func (f *Runtime) Edges(input *apipb.Filter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.FilterSearchEdges(input)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (f *Runtime) EdgesFrom(filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.RangeFilterFrom(filter)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (f *Runtime) EdgesTo(filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.RangeFilterTo(filter)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (r *Runtime) CreateNodes(nodes *apipb.NodeConstructors) (*apipb.Nodes, error) {
	for _, n := range nodes.GetNodes() {
		pathDefaults(n.Path)
	}
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_NodeConstructors{NodeConstructors: nodes},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	respNodes := resp.GetMutation().GetNodes()
	respNodes.Sort()
	return respNodes, nil
}

func (r *Runtime) CreateNode(node *apipb.NodeConstructor) (*apipb.Node, error) {
	pathDefaults(node.GetPath())
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_NODE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_NodeConstructor{NodeConstructor: node},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetNode(), nil
}

func (r *Runtime) PatchNodes(patches *apipb.Patches) (*apipb.Nodes, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patches{Patches: patches},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	respNodes := resp.GetMutation().GetNodes()
	respNodes.Sort()
	return respNodes, nil
}

func (r *Runtime) PatchNode(patch *apipb.Patch) (*apipb.Node, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_NODE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patch{Patch: patch},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetNode(), nil
}

func (r *Runtime) DelNodes(paths *apipb.Paths) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEmpty(), nil
}

func (r *Runtime) DelNode(path *apipb.Path) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEmpty(), nil
}

func (r *Runtime) CreateEdges(edges *apipb.EdgeConstructors) (*apipb.Edges, error) {
	for _, n := range edges.GetEdges() {
		pathDefaults(n.Path)
	}
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_EdgeConstructors{EdgeConstructors: edges},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	redges := resp.GetMutation().GetEdges()
	redges.Sort()
	return redges, nil
}

func (r *Runtime) CreateEdge(edge *apipb.EdgeConstructor) (*apipb.Edge, error) {
	pathDefaults(edge.Path)
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_EDGE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_EdgeConstructor{EdgeConstructor: edge},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEdge(), nil
}

func (r *Runtime) PatchEdges(patches *apipb.Patches) (*apipb.Edges, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patches{Patches: patches},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	redges := resp.GetMutation().GetEdges()
	redges.Sort()
	return redges, nil
}

func (r *Runtime) PatchEdge(patch *apipb.Patch) (*apipb.Edge, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_EDGE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patch{Patch: patch},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEdge(), nil
}

func (r *Runtime) DelEdges(paths *apipb.Paths) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEmpty(), nil
}

func (r *Runtime) DelEdge(path *apipb.Path) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.plugins {
		resp, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.plugins {
		change, err := plugin.HandleTrigger(context.Background(), &apipb.Trigger{
			Timing: apipb.Timing_AFTER,
			State:  resp,
		})
		if err != nil {
			return nil, err
		}
		resp = change
	}
	return resp.GetMutation().GetEmpty(), nil
}

func pathDefaults(path *apipb.Path) {
	if path == nil {
		path = &apipb.Path{}
	}
	if path.GetGid() == "" {
		path.Gid = uuid.New().String()
	}
	if path.GetGtype() == "" {
		path.Gtype = "default"
	}
}

func (r *Runtime) Export() (*apipb.Graph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nodes, err := r.graph.AllNodes()
	if err != nil {
		return nil, err
	}
	edges, err := r.graph.AllEdges()
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *Runtime) Import(graph *apipb.Graph) (*apipb.Graph, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes, err := r.graph.SetNodes(graph.GetNodes().GetNodes())
	if err != nil {
		return nil, err
	}
	edges, err := r.graph.SetEdges(graph.GetEdges().GetEdges())
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *Runtime) SubGraph(filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.SubGraph(filter)
}

func (r *Runtime) GetNodeDetail(filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.GetNodeDetail(filter)
}

func (r *Runtime) GetEdgeDetail(path *apipb.Path) (*apipb.EdgeDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.GetEdgeDetail(path)
}
