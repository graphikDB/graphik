package generic

import (
	"github.com/autom8ter/graphik/graph/model"
)

type NodeMap map[string]map[string]*model.Node

func (n NodeMap) Len(nodeType string) int {
	if c, ok := n[nodeType]; ok {
		return len(c)
	}
	return 0
}

func (n NodeMap) Types() []string {
	var nodeTypes []string
	for k, _ := range n {
		nodeTypes = append(nodeTypes, k)
	}
	return nodeTypes
}

func (n NodeMap) Get(nodeType string, id string) (*model.Node, bool) {
	if c, ok := n[nodeType]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n NodeMap) Set(value *model.Node) {
	if value.ID == "" {
		value.ID = uuid()
	}
	if _, ok := n[value.Type]; !ok {
		n[value.Type] = map[string]*model.Node{}
	}
	n[value.Type][value.ID] = value
}

func (n NodeMap) Range(nodeType string, f func(node *model.Node) bool) {
	if nodeType == Any {
		for _, c := range n {
			for _, v := range c {
				f(v)
			}
		}
	} else {
		if c, ok := n[nodeType]; ok {
			for _, v := range c {
				f(v)
			}
		}
	}
}

func (n NodeMap) Delete(nodeType string, id string) {
	if c, ok := n[nodeType]; ok {
		delete(c, id)
	}
}

func (n NodeMap) Exists(nodeType string, id string) bool {
	_, ok := n.Get(nodeType, id)
	return ok
}

func (n NodeMap) Filter(nodeType string, filter func(node *model.Node) bool) []*model.Node {
	var filtered []*model.Node
	n.Range(nodeType, func(node *model.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n NodeMap) SetAll(nodes ...*model.Node) {
	for _, node := range nodes {
		if _, ok := n[node.Type]; !ok {
			n[node.Type] = map[string]*model.Node{}
		}
		n[node.Type][node.ID] = node
	}
}

func (n NodeMap) DeleteAll(nodes ...*model.Node) {
	for _, node := range nodes {
		n.Delete(node.Type, node.ID)
	}
}

func (n NodeMap) Clear(nodeType string) {
	if cache, ok := n[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n NodeMap) Close() {
	for nodeType, _ := range n {
		n.Clear(nodeType)
	}
}
