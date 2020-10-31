package generic

import (
	"fmt"
	"github.com/autom8ter/graphik/graph/model"
	"github.com/jmespath/go-jmespath"
	"time"
)

type Nodes struct {
	nodes map[string]map[string]*model.Node
	edges *Edges
}

func NewNodes(edges *Edges) *Nodes {
	return &Nodes{
		nodes: map[string]map[string]*model.Node{},
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

func (n *Nodes) All() []*model.Node {
	var nodes []*model.Node
	n.Range(Any, func(node *model.Node) bool {
		nodes = append(nodes, node)
		return true
	})
	NodeList(nodes).Sort()
	return nodes
}

func (n *Nodes) Get(key model.Path) (*model.Node, bool) {
	if c, ok := n.nodes[key.Type]; ok {
		node := c[key.ID]
		return node, node != nil
	}
	return nil, false
}

func (n *Nodes) Set(value *model.Node) *model.Node {
	if value.Path.ID == "" {
		value.Path.ID = UUID()
	}
	if _, ok := n.nodes[value.Path.Type]; !ok {
		n.nodes[value.Path.Type] = map[string]*model.Node{}
	}
	n.nodes[value.Path.Type][value.Path.ID] = value
	return value
}

func (n *Nodes) Patch(updatedAt time.Time, value *model.Patch) *model.Node {
	if _, ok := n.nodes[value.Path.Type]; !ok {
		return nil
	}
	node := n.nodes[value.Path.Type][value.Path.ID]
	for k, v := range value.Patch {
		node.Attributes[k] = v
	}
	node.UpdatedAt = updatedAt
	return node
}

func (n *Nodes) Range(nodeType string, f func(node *model.Node) bool) {
	if nodeType == Any {
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

func (n *Nodes) Delete(key model.Path) bool {
	node, ok := n.Get(key)
	if !ok {
		return false
	}
	n.edges.RangeFrom(node.Path, func(e *model.Edge) bool {
		n.edges.Delete(e.Path)
		return true
	})
	n.edges.RangeTo(node.Path, func(e *model.Edge) bool {
		n.edges.Delete(e.Path)
		return true
	})
	if c, ok := n.nodes[key.Type]; ok {
		delete(c, key.ID)
	}
	return true
}

func (n *Nodes) Exists(key model.Path) bool {
	_, ok := n.Get(key)
	return ok
}

func (n *Nodes) Filter(nodeType string, filter func(node *model.Node) bool) []*model.Node {
	var filtered []*model.Node
	n.Range(nodeType, func(node *model.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	})
	NodeList(filtered).Sort()
	return filtered
}

func (n *Nodes) SetAll(nodes ...*model.Node) {
	for _, node := range nodes {
		n.Set(node)
	}
}

func (n *Nodes) DeleteAll(Nodes ...*model.Node) {
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

func (n *Nodes) Search(expression, nodeType string) (*model.SearchResults, error) {
	results := &model.SearchResults{
		Search: expression,
	}
	exp, err := jmespath.Compile(expression)
	if err != nil {
		return nil, err
	}
	n.Range(nodeType, func(node *model.Node) bool {
		val, _ := exp.Search(node)
		if val != nil {
			results.Results = append(results.Results, &model.SearchResult{
				Path: node.Path,
				Val:  val,
			})
		}
		return true
	})
	return results, nil
}

func (n *Nodes) FilterSearch(filter model.Filter) ([]*model.Node, error) {
	var nodes []*model.Node
	n.Range(filter.Type, func(node *model.Node) bool {
		for _, exp := range filter.Statements {
			val, _ := jmespath.Search(exp.Expression, node)
			if exp.Operator == model.OperatorNeq {
				if val == exp.Value {
					return true
				}
			}
			if exp.Operator == model.OperatorEq {
				if val != exp.Value {
					return true
				}
			}
		}
		nodes = append(nodes, node)
		return len(nodes) < filter.Limit
	})
	NodeList(nodes).Sort()
	return nodes, nil
}

func (n *Nodes) RangeFromDepth(filter model.DepthFilter, fn func(node *model.Node) bool) error {
	node, ok := n.Get(filter.Path)
	if !ok {
		return fmt.Errorf("%s does not exist", filter.Path.String())
	}
	walker := n.recurseFrom(0, filter, fn)
	walker(node)
	return nil
}

func (n *Nodes) RangeToDepth(filter model.DepthFilter, fn func(node *model.Node) bool) error {
	node, ok := n.Get(filter.Path)
	if !ok {
		return fmt.Errorf("%s does not exist", filter.Path.String())
	}
	walker := n.recurseTo(0, filter, fn)
	walker(node)
	return nil
}

func (n *Nodes) ascendFrom(seen map[string]struct{}, filter model.DepthFilter, fn func(node *model.Node) bool) func(node *model.Node) bool {
	return func(node *model.Node) bool {
		n.edges.RangeFrom(node.Path, func(e *model.Edge) bool {
			if e.Path.Type != filter.EdgeType {
				return true
			}
			node, ok := n.Get(e.To)
			if ok {
				if _, ok := seen[node.Path.String()]; !ok {
					seen[node.Path.String()] = struct{}{}
					for _, exp := range filter.Statements {
						val, _ := jmespath.Search(exp.Expression, node)
						if exp.Operator == model.OperatorNeq {
							if val == exp.Value {
								return true
							}
						}
						if exp.Operator == model.OperatorEq {
							if val != exp.Value {
								return true
							}
						}
					}
					return fn(node)
				}
			}
			return true
		})
		return true
	}
}

func (n *Nodes) ascendTo(seen map[string]struct{}, filter model.DepthFilter, fn func(node *model.Node) bool) func(node *model.Node) bool {
	return func(node *model.Node) bool {
		n.edges.RangeTo(node.Path, func(e *model.Edge) bool {
			if e.Path.Type != filter.EdgeType {
				return true
			}
			node, ok := n.Get(e.From)
			if ok {
				if _, ok := seen[node.Path.String()]; !ok {
					seen[node.Path.String()] = struct{}{}
					for _, exp := range filter.Statements {
						val, _ := jmespath.Search(exp.Expression, node)
						if exp.Operator == model.OperatorNeq {
							if val == exp.Value {
								return true
							}
						}
						if exp.Operator == model.OperatorEq {
							if val != exp.Value {
								return true
							}
						}
					}
					return fn(node)
				}
			}
			return true
		})
		return true
	}
}

func (n *Nodes) recurseFrom(depth int, filter model.DepthFilter, fn func(node *model.Node) bool) func(node *model.Node) bool {
	if depth > filter.Depth {
		return fn
	}
	depth += 1
	return n.recurseFrom(depth, filter, n.ascendFrom(map[string]struct{}{}, filter, fn))
}

func (n *Nodes) recurseTo(depth int, filter model.DepthFilter, fn func(node *model.Node) bool) func(node *model.Node) bool {
	if depth > filter.Depth {
		return fn
	}
	depth += 1
	return n.recurseTo(depth, filter, n.ascendTo(map[string]struct{}{}, filter, fn))
}

type NodeList []*model.Node

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
			return n[i].UpdatedAt.UnixNano() > n[j].UpdatedAt.UnixNano()
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
