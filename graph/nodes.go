package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type NodeStore struct {
	nodes map[string]map[string]*structpb.Struct
	edges *EdgeStore
}

func newNodeStore(edges *EdgeStore) *NodeStore {
	return &NodeStore{
		nodes: map[string]map[string]*structpb.Struct{},
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

func (n *NodeStore) All() *apipb.Structs {
	var nodes []*structpb.Struct
	n.Range(apipb.Keyword_ANY.String(), func(node *structpb.Struct) bool {
		nodes = append(nodes, node)
		return true
	})
	return &apipb.Structs{
		Values: nodes,
	}
}

func (n *NodeStore) Get(path string) (*structpb.Struct, bool) {
	xtype, xid := lang.SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		node := c[xid]
		return node, node != nil
	}
	return nil, false
}

func (n *NodeStore) Set(value *structpb.Struct) *structpb.Struct {
	if _, ok := n.nodes[lang.GetType(value)]; !ok {
		n.nodes[lang.GetType(value)] = map[string]*structpb.Struct{}
	}
	n.nodes[lang.GetType(value)][lang.GetID(value)] = value
	return value
}

func (n *NodeStore) Patch(updatedAt *timestamp.Timestamp, value *apipb.Patch) *structpb.Struct {
	xtype, xid := lang.SplitPath(value.Path)
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

func (n *NodeStore) Range(nodeType string, f func(node *structpb.Struct) bool) {
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
	xtype, xid := lang.SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		delete(c, xid)
	}
	return true
}

func (n *NodeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (n *NodeStore) Filter(nodeType string, filter func(node *structpb.Struct) bool) *apipb.Structs {
	var filtered []*structpb.Struct
	n.Range(nodeType, func(node *structpb.Struct) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return &apipb.Structs{
		Values: filtered,
	}
}

func (n *NodeStore) SetAll(nodes *apipb.Structs) {
	for _, node := range nodes.Values {
		n.Set(node)
	}
}

func (n *NodeStore) DeleteAll(nodes *apipb.Structs) {
	for _, node := range nodes.Values {
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

func (n *NodeStore) FilterSearch(filter *apipb.Filter) (*apipb.Structs, error) {
	var nodes []*structpb.Struct
	var err error
	var pass bool
	n.Range(filter.Type, func(node *structpb.Struct) bool {
		pass, err = lang.BooleanExpression(filter.Expressions, node)
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
