package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
)

type NodeStore struct {
	nodes map[string]map[string]*lang.Values
	edges *EdgeStore
}

func newNodeStore(edges *EdgeStore) *NodeStore {
	return &NodeStore{
		nodes: map[string]map[string]*lang.Values{},
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

func (n *NodeStore) All() []*lang.Values {
	var nodes []*lang.Values
	n.Range(apipb.Keyword_ANY.String(), func(node *lang.Values) bool {
		nodes = append(nodes, node)
		return true
	})
	return nodes
}

func (n *NodeStore) Get(path string) (*lang.Values, bool) {
	xtype, xid := lang.SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		node := c[xid]
		return node, node != nil
	}
	return nil, false
}

func (n *NodeStore) Set(value *lang.Values) *lang.Values {
	if _, ok := n.nodes[value.GetType()]; !ok {
		n.nodes[value.GetType()] = map[string]*lang.Values{}
	}
	n.nodes[value.GetType()][value.GetID()] = value
	return value
}

func (n *NodeStore) Patch(updatedAt int64, value *lang.Values) *lang.Values {
	if _, ok := n.nodes[value.GetType()]; !ok {
		return nil
	}
	node := n.nodes[value.GetType()][value.GetID()]
	for k, v := range value.Fields {
		node.Fields[k] = v
	}
	node.Fields["updated_at"] = lang.ToValue(updatedAt)
	return node
}

func (n *NodeStore) Range(nodeType string, f func(node *lang.Values) bool) {
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
	n.edges.RangeFrom(path, func(e *lang.Values) bool {
		n.edges.Delete(e.PathString())
		if e.GetString("cascade") == apipb.Cascade_TO.String() || e.GetString("cascade") == apipb.Cascade_MUTUAL.String() {
			n.Delete(e.GetString("to"))
		}
		return true
	})
	n.edges.RangeTo(path, func(e *lang.Values) bool {
		n.edges.Delete(e.PathString())
		if e.GetString("cascade") == apipb.Cascade_FROM.String() || e.GetString("cascade") == apipb.Cascade_MUTUAL.String() {
			n.Delete(e.GetString("from"))
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

func (n *NodeStore) Filter(nodeType string, filter func(node *lang.Values) bool) []*lang.Values {
	var filtered []*lang.Values
	n.Range(nodeType, func(node *lang.Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n *NodeStore) SetAll(nodes []*lang.Values) {
	for _, node := range nodes {
		n.Set(node)
	}
}

func (n *NodeStore) DeleteAll(nodes []*lang.Values) {
	for _, node := range nodes {
		n.Delete(node.PathString())
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

func (n *NodeStore) FilterSearch(filter *apipb.Filter) ([]*lang.Values, error) {
	var nodes []*lang.Values
	var err error
	var pass bool
	n.Range(filter.Type, func(node *lang.Values) bool {
		pass, err = lang.BooleanExpression(filter.Expressions, node)
		if err != nil {
			return false
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})
	return nodes, err
}
