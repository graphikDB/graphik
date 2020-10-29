package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"

	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/graph/model"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input model.NodeConstructor) (*model.Node, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.CREATE_NODE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) PatchNode(ctx context.Context, input *model.Patch) (*model.Node, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.PATCH_NODE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) DelNode(ctx context.Context, input model.ForeignKey) (*model.Counter, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.DELETE_NODE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Counter), nil
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input model.EdgeConstructor) (*model.Edge, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.CREATE_EDGE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) PatchEdge(ctx context.Context, input model.Patch) (*model.Edge, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.PATCH_EDGE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) DelEdge(ctx context.Context, input model.ForeignKey) (*model.Counter, error) {
	res, err := r.store.Execute(&command.Command{
		Op:  command.DELETE_EDGE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Counter), nil
}

func (r *queryResolver) Node(ctx context.Context, input model.ForeignKey) (*model.Node, error) {
	key := &dagger.ForeignKey{
		XID:   input.ID,
		XType: input.Type,
	}
	n, ok := dagger.GetNode(key)
	if !ok {
		return nil, fmt.Errorf("node not found: %s", key.Path())
	}
	return &model.Node{
		Attributes: n.Raw(),
		Edges:      nil,
	}, nil
}

func (r *queryResolver) Nodes(ctx context.Context, input model.NodeFilter) ([]*model.Node, error) {
	var nodes []*model.Node
	dagger.RangeNodeTypes(dagger.StringType(input.Type), func(n *dagger.Node) bool {
		for _, filter := range input.Expressions {
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
			Attributes: n.Raw(),
			Edges:      nil,
		}
		nodes = append(nodes, node)
		return len(nodes) < input.Limit
	})
	return nodes, nil
}

func (r *queryResolver) Edge(ctx context.Context, input model.ForeignKey) (*model.Edge, error) {
	key := &dagger.ForeignKey{
		XID:   input.ID,
		XType: input.Type,
	}
	e, ok := dagger.GetEdge(key)
	if !ok {
		return nil, fmt.Errorf("edge not found: %s", key.Path())
	}
	return &model.Edge{
		Attributes: e.Node().Raw(),
		From: &model.Node{
			Attributes: e.From().Raw(),
			Edges:      nil,
		},
		To: &model.Node{
			Attributes: e.To().Raw(),
			Edges:      nil,
		},
	}, nil
}

func (r *queryResolver) Edges(ctx context.Context, input model.EdgeFilter) ([]*model.Edge, error) {
	var edges []*model.Edge
	dagger.RangeEdgeTypes(dagger.StringType(input.Type), func(edge *dagger.Edge) bool {
		for _, filter := range input.Expressions {
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
			} else if strings.Contains(filter.Key, "to.") {
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
			Attributes: edge.Node().Raw(),
			From: &model.Node{
				Attributes: edge.From().Raw(),
				Edges:      nil,
			},
			To: &model.Node{
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
