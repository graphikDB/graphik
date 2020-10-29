package generic

import (
	"github.com/autom8ter/graphik/graph/model"
)

type Nodes struct {
	nodes map[string]map[string]*model.Node
}


func NewNodes() *Nodes {
	return &Nodes{
		nodes: map[string]map[string]*model.Node{},
	}
}

func (n Nodes) Len(nodeType string) int {
	if c, ok := n.nodes[nodeType]; ok {
		return len(c)
	}
	return 0
}

func (n Nodes) Types() []string {
	var nodeTypes []string
	for k, _ := range n.nodes {
		nodeTypes = append(nodeTypes, k)
	}
	return nodeTypes
}

func (n Nodes) All() []*model.Node {
	var nodes []*model.Node
	n.Range(Any, func(node *model.Node) bool {
		nodes = append(nodes, node)
		return true
	})
	return nodes
}

func (n Nodes) Get(nodeType string, id string) (*model.Node, bool) {
	if c, ok := n.nodes[nodeType]; ok {
		node := c[id]
		return c[id], node != nil
	}
	return nil, false
}

func (n Nodes) Set(value *model.Node) {
	if value.ID == "" {
		value.ID = uuid()
	}
	if _, ok := n.nodes[value.Type]; !ok {
		n.nodes[value.Type] = map[string]*model.Node{}
	}
	n.nodes[value.Type][value.ID] = value
}

func (n Nodes) Range(nodeType string, f func(node *model.Node) bool) {
	if nodeType == Any {
		for _, c := range n.nodes {
			for _, v := range c {
				f(v)
			}
		}
	} else {
		if c, ok := n.nodes[nodeType]; ok {
			for _, v := range c {
				f(v)
			}
		}
	}
}

func (n Nodes) Delete(nodeType string, id string) {
	if c, ok := n.nodes[nodeType]; ok {
		delete(c, id)
	}
}

func (n Nodes) Exists(nodeType string, id string) bool {
	_, ok := n.Get(nodeType, id)
	return ok
}

func (n Nodes) Filter(nodeType string, filter func(node *model.Node) bool) []*model.Node {
	var filtered []*model.Node
	n.Range(nodeType, func(node *model.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	return filtered
}

func (n Nodes) SetAll(Nodes ...*model.Node) {
	for _, node := range Nodes {
		if _, ok := n.nodes[node.Type]; !ok {
			n.nodes[node.Type] = map[string]*model.Node{}
		}
		n.nodes[node.Type][node.ID] = node
	}
}

func (n Nodes) DeleteAll(Nodes ...*model.Node) {
	for _, node := range Nodes {
		n.Delete(node.Type, node.ID)
	}
}

func (n Nodes) Clear(nodeType string) {
	if cache, ok := n.nodes[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n Nodes) Close() {
	for nodeType, _ := range n.nodes {
		n.Clear(nodeType)
	}
}
