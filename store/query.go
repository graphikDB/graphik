package store

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/graph/model"
	"strings"
)

func (f *Store) Node(ctx context.Context, input model.ForeignKey) (*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	node, ok := f.nodes.Get(input)
	if !ok {
		return nil, fmt.Errorf("node %s.%s does not exist", input.Type, input.ID)
	}
	return node, nil
}

func (f *Store) Nodes(ctx context.Context, input model.NodeFilter) ([]*model.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var nodes []*model.Node
	f.nodes.Range(input.Type, func(n *model.Node) bool {
		for _, filter := range input.Expressions {
			if filter.Operator == "!=" {
				if n.Attributes[filter.Key] == filter.Value {
					return true
				}
			}
			if filter.Operator == "==" {
				if n.Attributes[filter.Key] != filter.Value {
					return true
				}
			}
		}
		nodes = append(nodes, n)
		return len(nodes) < input.Limit
	})
	return nodes, nil
}

func (f *Store) Edge(ctx context.Context, input model.ForeignKey) (*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	edge, ok := f.edges.Get(input)
	if !ok {
		return nil, fmt.Errorf("edge node %s.%s does not exist", input.Type, input.ID)
	}
	return edge, nil
}

func (f *Store) Edges(ctx context.Context, input model.EdgeFilter) ([]*model.Edge, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var edges []*model.Edge
	f.edges.Range(input.Type, func(edge *model.Edge) bool {
		for _, filter := range input.Expressions {
			if strings.Contains(filter.Key, "from.") {
				split := strings.Split(filter.Key, "from.")
				if len(split) > 1 {
					if filter.Operator == "!=" {
						if edge.From.Attributes[split[1]] == filter.Value {
							return true
						}
					}
					if filter.Operator == "==" {
						if edge.From.Attributes[split[1]] != filter.Value {
							return true
						}
					}
				}
			} else if strings.Contains(filter.Key, "to.") {
				split := strings.Split(filter.Key, "to.")
				if len(split) > 1 {
					if filter.Operator == "!=" {
						if edge.To.Attributes[split[1]] == filter.Value {
							return true
						}
					}
					if filter.Operator == "==" {
						if edge.To.Attributes[split[1]] != filter.Value {
							return true
						}
					}
				}
			} else {
				if filter.Operator == "!=" {
					if edge.Attributes[filter.Key] == filter.Value {
						return true
					}
				}
				if filter.Operator == "==" {
					if edge.Attributes[filter.Key] != filter.Value {
						return true
					}
				}
			}
		}
		edges = append(edges, edge)
		return len(edges) < input.Limit
	})
	return edges, nil
}
