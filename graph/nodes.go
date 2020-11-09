package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type NodeStore struct {
	nodes map[string]map[string]*apipb.Node
	edges *EdgeStore
}

func newNodeStore(edges *EdgeStore) *NodeStore {
	return &NodeStore{
		nodes: map[string]map[string]*apipb.Node{},
		edges: edges,
	}
}

func (n *NodeStore) Len(nodeType string) int {
	if c, ok := n.nodes[nodeType]; ok {
		return len(c)
	}
	return 0
}

func (n *NodeStore) Types() []string {
	var nodeTypes []string
	for k, _ := range n.nodes {
		nodeTypes = append(nodeTypes, k)
	}
	return nodeTypes
}

func (n *NodeStore) All() *apipb.Nodes {
	var nodes []*apipb.Node
	n.Range(apipb.Keyword_ANY.String(), func(node *apipb.Node) bool {
		nodes = append(nodes, node)
		return true
	})
	return &apipb.Nodes{
		Nodes: nodes,
	}
}

func (n *NodeStore) Get(path string) (*apipb.Node, bool) {
	xtype, xid := apipb.SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		node := c[xid]
		return node, node != nil
	}
	return nil, false
}

func (n *NodeStore) Set(value *apipb.Node) *apipb.Node {
	xtype, xid := apipb.SplitPath(value.Path)

	if _, ok := n.nodes[xtype]; !ok {
		n.nodes[xtype] = map[string]*apipb.Node{}
	}
	n.nodes[xtype][xid] = value
	return value
}

func (n *NodeStore) Patch(updatedAt *timestamp.Timestamp, value *apipb.Patch) *apipb.Node {
	xtype, xid := apipb.SplitPath(value.Path)
	if _, ok := n.nodes[xtype]; !ok {
		return nil
	}
	node := n.nodes[xtype][xid]
	for k, v := range value.Patch.Fields {
		node.Attributes.Fields[k] = v
	}
	node.UpdatedAt = updatedAt
	return node
}

func (n *NodeStore) Range(nodeType string, f func(node *apipb.Node) bool) {
	if nodeType == apipb.Keyword_ANY.String() {
		for _, c := range n.nodes {
			for _, node := range c {
				f(node)
			}
		}
	} else {
		if c, ok := n.nodes[nodeType]; ok {
			for _, node := range c {
				f(node)
			}
		}
	}
}

func (n *NodeStore) Delete(path string) bool {
	node, ok := n.Get(path)
	if !ok {
		return false
	}
	n.edges.RangeFrom(node.Path, func(e *apipb.Edge) bool {
		n.edges.Delete(e.Path)
		if e.Cascade == apipb.Cascade_TO || e.Cascade == apipb.Cascade_MUTUAL {
			n.Delete(e.To)
		}
		return true
	})
	n.edges.RangeTo(node.Path, func(e *apipb.Edge) bool {
		n.edges.Delete(e.Path)
		if e.Cascade == apipb.Cascade_FROM || e.Cascade == apipb.Cascade_MUTUAL {
			n.Delete(e.From)
		}
		return true
	})
	xtype, xid := apipb.SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		delete(c, xid)
	}
	return true
}

func (n *NodeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (n *NodeStore) Filter(nodeType string, filter func(node *apipb.Node) bool) *apipb.Nodes {
	var filtered []*apipb.Node
	n.Range(nodeType, func(node *apipb.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return &apipb.Nodes{
		Nodes: filtered,
	}
}

func (n *NodeStore) SetAll(nodes *apipb.Nodes) {
	for _, node := range nodes.Nodes {
		n.Set(node)
	}
}

func (n *NodeStore) DeleteAll(nodes *apipb.Nodes) {
	for _, node := range nodes.Nodes {
		n.Delete(node.Path)
	}
}

func (n *NodeStore) Clear(nodeType string) {
	if cache, ok := n.nodes[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n *NodeStore) Close() {
	for nodeType, _ := range n.nodes {
		n.Clear(nodeType)
	}
}

func (n *NodeStore) FilterSearch(filter *apipb.Filter) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	var err error
	var pass bool
	n.Range(filter.Type, func(node *apipb.Node) bool {
		pass, err = apipb.EvaluateExpressions(filter.Expressions, node)
		if err != nil {
			return false
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})
	return &apipb.Nodes{
		Nodes: nodes,
	}, err
}
