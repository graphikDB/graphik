package generic

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

type Nodes struct {
	nodes map[string]map[string]*apipb.Node
	edges *Edges
}

func NewNodes(edges *Edges) *Nodes {
	return &Nodes{
		nodes: map[string]map[string]*apipb.Node{},
		edges: edges,
	}
}

func (n *Nodes) Len(nodeType string) int {
	if c, ok := n.nodes[nodeType]; ok {
		return len(c)
	}
	return 0
}

func (n *Nodes) Types() []string {
	var nodeTypes []string
	for k, _ := range n.nodes {
		nodeTypes = append(nodeTypes, k)
	}
	return nodeTypes
}

func (n *Nodes) All() []*apipb.Node {
	var nodes []*apipb.Node
	n.Range(apipb.Keyword_ANY.String(), func(node *apipb.Node) bool {
		nodes = append(nodes, node)
		return true
	})
	NodeList(nodes).Sort()
	return nodes
}

func (n *Nodes) Get(path *apipb.Path) (*apipb.Node, bool) {
	if c, ok := n.nodes[path.Type]; ok {
		node := c[path.ID]
		return node, node != nil
	}
	return nil, false
}

func (n *Nodes) Set(value *apipb.Node) *apipb.Node {
	if value.Path.ID == "" {
		value.Path.ID = UUID()
	}
	if _, ok := n.nodes[value.Path.Type]; !ok {
		n.nodes[value.Path.Type] = map[string]*apipb.Node{}
	}
	n.nodes[value.Path.Type][value.Path.ID] = value
	return value
}

func (n *Nodes) Patch(updatedAt *timestamp.Timestamp, value *apipb.Patch) *apipb.Node {
	if _, ok := n.nodes[value.Path.Type]; !ok {
		return nil
	}
	node := n.nodes[value.Path.Type][value.Path.ID]
	for k, v := range value.Patch.Fields {
		node.Attributes.Fields[k] = v
	}
	node.UpdatedAt = updatedAt
	return node
}

func (n *Nodes) Range(nodeType string, f func(node *apipb.Node) bool) {
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

func (n *Nodes) Delete(path *apipb.Path) bool {
	node, ok := n.Get(path)
	if !ok {
		return false
	}
	n.edges.RangeFrom(node.Path, func(e *apipb.Edge) bool {
		n.edges.Delete(e.Path)
		return true
	})
	n.edges.RangeTo(node.Path, func(e *apipb.Edge) bool {
		n.edges.Delete(e.Path)
		return true
	})
	if c, ok := n.nodes[path.Type]; ok {
		delete(c, path.ID)
	}
	return true
}

func (n *Nodes) Exists(path *apipb.Path) bool {
	_, ok := n.Get(path)
	return ok
}

func (n *Nodes) Filter(nodeType string, filter func(node *apipb.Node) bool) []*apipb.Node {
	var filtered []*apipb.Node
	n.Range(nodeType, func(node *apipb.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	NodeList(filtered).Sort()
	return filtered
}

func (n *Nodes) SetAll(nodes ...*apipb.Node) {
	for _, node := range nodes {
		n.Set(node)
	}
}

func (n *Nodes) DeleteAll(Nodes ...*apipb.Node) {
	for _, node := range Nodes {
		n.Delete(node.Path)
	}
}

func (n *Nodes) Clear(nodeType string) {
	if cache, ok := n.nodes[nodeType]; ok {
		for k, _ := range cache {
			delete(cache, k)
		}
	}
}

func (n *Nodes) Close() {
	for nodeType, _ := range n.nodes {
		n.Clear(nodeType)
	}
}

func (n *Nodes) FilterSearch(filter apipb.Filter) ([]*apipb.Node, error) {
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
	NodeList(nodes).Sort()
	return nodes, err
}

type NodeList []*apipb.Node

func (n NodeList) Sort() {
	sorter := Interface{
		LenFunc: func() int {
			if n == nil {
				return 0
			}
			return len(n)
		},
		LessFunc: func(i, j int) bool {
			if n == nil {
				return false
			}
			return n[i].UpdatedAt.Nanos > n[j].UpdatedAt.Nanos
		},
		SwapFunc: func(i, j int) {
			if n == nil {
				return
			}
			n[i], n[j] = n[j], n[i]
		},
	}
	sorter.Sort()
}
