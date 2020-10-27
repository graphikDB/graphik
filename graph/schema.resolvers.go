package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/primitive"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/graph/model"
	"strings"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input model.NodeInput) (*model.Node, error) {
	node := dagger.NewNode(input.Type, "", input.Attributes)
	return &model.Node{
		ID:         node.ID(),
		Type:       node.Type(),
		Attributes: node.Raw(),
	}, nil
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input model.EdgeInput) (*model.Edge, error) {
	from, ok := dagger.GetNode(dagger.ForeignKey(input.From.Type, input.From.ID))
	if !ok {
		return nil, fmt.Errorf("%s.%s does not exist", input.From.Type, input.From.ID)
	}
	to, ok := dagger.GetNode(dagger.ForeignKey(input.To.Type, input.To.ID))
	if !ok {
		return nil, fmt.Errorf("%s.%s does not exist", input.To.Type, input.To.ID)
	}
	edge, err := dagger.NewEdge(input.Node.Type, nullString(input.Node.ID), input.Node.Attributes, from, to)
	if err != nil {
		return nil, err
	}
	return &model.Edge{
		Node: &model.Node{
			ID:         edge.ID(),
			Type:       edge.Type(),
			Attributes: nil,
			Edges:      nil,
		},
		From: &model.Node{
			ID:         from.ID(),
			Type:       from.Type(),
			Attributes: from.Raw(),
			Edges:      nil,
		},
		To: &model.Node{
			ID:         to.ID(),
			Type:       to.Type(),
			Attributes: to.Raw(),
			Edges:      nil,
		},
	}, nil
}

func (r *queryResolver) Nodes(ctx context.Context, input model.QueryNodes) ([]*model.Node, error) {
	var nodes []*model.Node
	dagger.RangeNodeTypes(primitive.StringType(input.Type), func(n *dagger.Node) bool {
		for _, filter := range input.Filters {
			if filter.Operator == "!=" {
				if n.Get(filter.Key) == filter.Value {
					return true
				}
			}
			if filter.Operator == "==" {
				if n.Get(filter.Key) != filter.Value {
					return true
				}
			}
		}
		node := &model.Node{
			ID:         n.ID(),
			Type:       n.Type(),
			Attributes: n.Raw(),
		}
		nodes = append(nodes, node)
		return len(nodes) < input.Limit
	})
	return nodes, nil
}

func (r *queryResolver) Edges(ctx context.Context, input model.QueryEdges) ([]*model.Edge, error) {
	var edges []*model.Edge
	dagger.RangeEdgeTypes(primitive.StringType(input.Type), func(edge *dagger.Edge) bool {
		for _, filter := range input.Filters {
			if strings.Contains(filter.Key, "from.") {
				split := strings.Split(filter.Key, "from.")
				if len(split) > 1 {
					if filter.Operator == "!=" {
						if edge.From().Get(split[1]) == filter.Value {
							return true
						}
					}
					if filter.Operator == "==" {
						if edge.From().Get(split[1]) != filter.Value {
							return true
						}
					}
				}
			}else if strings.Contains(filter.Key, "to.") {
				split := strings.Split(filter.Key, "to.")
				if len(split) > 1 {
					if filter.Operator == "!=" {
						if edge.To().Get(split[1]) == filter.Value {
							return true
						}
					}
					if filter.Operator == "==" {
						if edge.To().Get(split[1]) != filter.Value {
							return true
						}
					}
				}
			} else {
				if filter.Operator == "!=" {
					if edge.Get(filter.Key) == filter.Value {
						return true
					}
				}
				if filter.Operator == "==" {
					if edge.Get(filter.Key) != filter.Value {
						return true
					}
				}
			}
		}
		edges = append(edges, &model.Edge{
			Node: &model.Node{
				ID:         edge.ID(),
				Type:       edge.Type(),
				Attributes: nil,
				Edges:      nil,
			},
			From: &model.Node{
				ID:         edge.From().ID(),
				Type:       edge.From().Type(),
				Attributes: edge.From().Raw(),
				Edges:      nil,
			},
			To: &model.Node{
				ID:         edge.To().ID(),
				Type:       edge.To().Type(),
				Attributes: edge.To().Raw(),
				Edges:      nil,
			},
		})
		return len(edges) < input.Limit
	})
	return edges, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func nullString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
