package graph

import (
	"sort"
	"time"
)

type NodeStore struct {
	nodes map[string]map[string]Values
	edges *EdgeStore
}

func newNodeStore(edges *EdgeStore) *NodeStore {
	return &NodeStore{
		nodes: map[string]map[string]Values{},
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
	sort.Strings(nodeTypes)
	return nodeTypes
}

func (n *NodeStore) All() ValueSet {
	var nodes ValueSet
	n.Range(Any, func(node Values) bool {
		nodes = append(nodes, node)
		return true
	})
	nodes.Sort()
	return nodes
}

func (n *NodeStore) Get(path string) (Values, bool) {
	xtype, xid := SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		node := c[xid]
		return node, node != nil
	}
	return nil, false
}

func (n *NodeStore) Set(value Values) Values {
	if value.GetID() == "" {
		value.SetID(UUID())
	}
	if value.GetType() == "" {
		value.SetType(Default)
	}
	if value.GetCreatedAt() == 0 {
		value.SetCreatedAt(time.Now())
	}
	if value.GetUpdatedAt() == 0 {
		value.SetUpdatedAt(time.Now())
	}
	if _, ok := n.nodes[value.GetType()]; !ok {
		n.nodes[value.GetType()] = map[string]Values{}
	}
	n.nodes[value.GetType()][value.GetID()] = value
	return value
}

func (n *NodeStore) Patch(value Values) Values {
	value.SetUpdatedAt(time.Now())
	if _, ok := n.nodes[value.GetType()]; !ok {
		return nil
	}
	node := n.nodes[value.GetType()][value.GetID()]
	for k, v := range value {
		node[k] = v
	}
	return node
}

func (n *NodeStore) Range(nodeType string, f func(node Values) bool) {
	if nodeType == Any {
		for _, c := range n.nodes {
			for _, node := range c {
				if !f(node) {
					return
				}
			}
		}
	} else {
		if c, ok := n.nodes[nodeType]; ok {
			for _, node := range c {
				if !f(node) {
					return
				}
			}
		}
	}
}

func (n *NodeStore) Delete(path string) bool {
	if !n.Exists(path) {
		return false
	}
	n.edges.RangeFrom(path, func(e Values) bool {
		n.edges.Delete(e.GetPath())
		if this := e.GetString(CascadeKey); this != "" {
			if this == CascadeTo || this == CascadeMutual {
				n.Delete(e.GetString(ToKey))
			}
		}
		return true
	})
	n.edges.RangeTo(path, func(e Values) bool {
		n.edges.Delete(e.GetPath())
		if this := e.GetString(CascadeKey); this != "" {
			if this == CascadeFrom || this == CascadeMutual {
				n.Delete(e.GetString(FromKey))
			}
		}
		return true
	})
	xtype, xid := SplitPath(path)
	if c, ok := n.nodes[xtype]; ok {
		delete(c, xid)
	}
	return true
}

func (n *NodeStore) Exists(path string) bool {
	_, ok := n.Get(path)
	return ok
}

func (n *NodeStore) Filter(nodeType string, filter func(node Values) bool) ValueSet {
	var filtered ValueSet
	n.Range(nodeType, func(node Values) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	filtered.Sort()
	return filtered
}

func (n *NodeStore) SetAll(nodes ValueSet) {
	for _, node := range nodes {
		n.Set(node)
	}
}

func (n *NodeStore) DeleteAll(nodes ValueSet) {
	for _, node := range nodes {
		n.Delete(node.GetPath())
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

func (n *NodeStore) FilterSearch(filter *Filter) (ValueSet, error) {
	var nodes ValueSet
	var err error
	var pass bool
	n.Range(filter.Type, func(node Values) bool {
		pass, err = BooleanExpression(filter.Expressions, node)
		if err != nil {
			return false
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	})
	nodes.Sort()
	return nodes, err
}
