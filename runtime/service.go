package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"time"
)

func (f *Runtime) Node(ctx context.Context, input *apipb.Path) (*apipb.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.GetNode(ctx, input)
}

func (f *Runtime) Nodes(ctx context.Context, input *apipb.Filter) (*apipb.Nodes, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	nodes, err := f.graph.FilterSearchNodes(ctx, input)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return nodes, nil
}

func (f *Runtime) Edge(ctx context.Context, input *apipb.Path) (*apipb.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.graph.GetEdge(ctx, input)
}

func (f *Runtime) Edges(ctx context.Context, input *apipb.Filter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.FilterSearchEdges(ctx, input)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (f *Runtime) EdgesFrom(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.RangeFilterFrom(ctx, filter)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (f *Runtime) EdgesTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edges, err := f.graph.RangeFilterTo(ctx, filter)
	if err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
}

func (r *Runtime) CreateNodes(ctx context.Context, nodes *apipb.NodeConstructors) (*apipb.Nodes, error) {
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
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) CreateNode(ctx context.Context, node *apipb.NodeConstructor) (*apipb.Node, error) {
	pathDefaults(node.GetPath())
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_NODE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_NodeConstructor{NodeConstructor: node},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) PatchNodes(ctx context.Context, patches *apipb.Patches) (*apipb.Nodes, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patches{Patches: patches},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) PatchNode(ctx context.Context, patch *apipb.Patch) (*apipb.Node, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_NODE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patch{Patch: patch},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) DelNodes(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) DelNode(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_NODES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) CreateEdges(ctx context.Context, edges *apipb.EdgeConstructors) (*apipb.Edges, error) {
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
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) CreateEdge(ctx context.Context, edge *apipb.EdgeConstructor) (*apipb.Edge, error) {
	pathDefaults(edge.Path)
	change := &apipb.StateChange{
		Op: apipb.Op_CREATE_EDGE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_EdgeConstructor{EdgeConstructor: edge},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) PatchEdges(ctx context.Context, patches *apipb.Patches) (*apipb.Edges, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patches{Patches: patches},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) PatchEdge(ctx context.Context, patch *apipb.Patch) (*apipb.Edge, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_PATCH_EDGE,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Patch{Patch: patch},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) DelEdges(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Paths{Paths: paths},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) DelEdge(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	change := &apipb.StateChange{
		Op: apipb.Op_DELETE_EDGES,
		Mutation: &apipb.Mutation{
			Object: &apipb.Mutation_Path{Path: path},
		},
		Timestamp: time.Now().UnixNano(),
	}
	for _, plugin := range r.triggers {
		resp, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
			Timing: apipb.Timing_BEFORE,
			State:  change,
		})
		if err != nil {
			return nil, err
		}
		change = resp
	}
	resp, err := r.execute(ctx, change)
	if err != nil {
		return nil, err
	}
	for _, plugin := range r.triggers {
		change, err := plugin.HandleTrigger(ctx, &apipb.Trigger{
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

func (r *Runtime) Export(ctx context.Context) (*apipb.Graph, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nodes, err := r.graph.AllNodes(ctx)
	if err != nil {
		return nil, err
	}
	edges, err := r.graph.AllEdges(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *Runtime) Import(ctx context.Context, graph *apipb.Graph) (*apipb.Graph, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes, err := r.graph.SetNodes(graph.GetNodes().GetNodes())
	if err != nil {
		return nil, err
	}
	edges, err := r.graph.SetEdges(ctx, graph.GetEdges().GetEdges())
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *Runtime) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.SubGraph(ctx, filter)
}

func (r *Runtime) GetNodeDetail(ctx context.Context, filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.GetNodeDetail(ctx, filter)
}

func (r *Runtime) GetEdgeDetail(ctx context.Context, path *apipb.Path) (*apipb.EdgeDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.GetEdgeDetail(ctx, path)
}

func (r *Runtime) Schema(ctx context.Context) (*apipb.Schema, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.graph.Schema(ctx)
}
