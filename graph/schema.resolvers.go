package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/autom8ter/dagger"
	"github.com/autom8ter/dagger/primitive"
	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/graph/model"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input map[string]interface{}) (*model.Node, error) {
	if input[primitive.ID_KEY] == nil {
		input[primitive.ID_KEY] = dagger.RandomID().ID()
	}
	if input[primitive.TYPE_KEY] == nil {
		input[primitive.TYPE_KEY] = primitive.DefaultType
	}
	res, err := r.store.Execute(&command.Command{
		Op:  command.CREATE_NODE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) SetNode(ctx context.Context, input map[string]interface{}) (*model.Node, error) {
	if input[primitive.ID_KEY] == nil {
		return nil, errors.New("emtpy node _id")
	}
	if input[primitive.TYPE_KEY] == nil {
		return nil, errors.New("emtpy node _type")
	}
	if !dagger.HasNode(&dagger.ForeignKey{
		XID:   input[primitive.ID_KEY].(string),
		XType: input[primitive.TYPE_KEY].(string),
	}) {
		return nil, errors.New("node does not exist")
	}
	res, err := r.store.Execute(&command.Command{
		Op:  command.SET_NODE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) DelNode(ctx context.Context, input model.ForeignKey) (*model.Counter, error) {
	key := &dagger.ForeignKey{
		XID:   input.ID,
		XType: input.Type,
	}
	if dagger.HasNode(key) {
		_, err := r.store.Execute(&command.Command{
			Op:  command.DELETE_NODE,
			Val: input,
		})
		if err != nil {
			return nil, err
		}
		return &model.Counter{Count: 1}, nil
	}
	return &model.Counter{
		Count: 0,
	}, nil
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input model.EdgeInput) (*model.Edge, error) {
	if input.Node[primitive.ID_KEY] == nil {
		input.Node[primitive.ID_KEY] = dagger.RandomID().ID()
	}
	if input.Node[primitive.TYPE_KEY] == nil {
		input.Node[primitive.TYPE_KEY] = primitive.DefaultType
	}
	res, err := r.store.Execute(&command.Command{
		Op:  command.CREATE_EDGE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) SetEdge(ctx context.Context, input model.EdgeInput) (*model.Edge, error) {
	if !dagger.HasEdge(&dagger.ForeignKey{
		XID:   input.Node[primitive.ID_KEY].(string),
		XType: input.Node[primitive.TYPE_KEY].(string),
	}) {
		return nil, errors.New("edge node does not exist")
	}
	res, err := r.store.Execute(&command.Command{
		Op:  command.SET_EDGE,
		Val: input,
	})
	if err != nil {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) DelEdge(ctx context.Context, input model.ForeignKey) (*model.Counter, error) {
	key := &dagger.ForeignKey{
		XID:   input.ID,
		XType: input.Type,
	}
	if dagger.HasEdge(key) {
		_, err := r.store.Execute(&command.Command{
			Op:  command.DELETE_EDGE,
			Val: input,
		})
		if err != nil {
			return nil, err
		}
		return &model.Counter{Count: 1}, nil
	}
	return &model.Counter{
		Count: 0,
	}, nil
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

func (r *queryResolver) Nodes(ctx context.Context, input model.QueryNodes) ([]*model.Node, error) {
	var nodes []*model.Node
	dagger.RangeNodeTypes(dagger.StringType(input.Type), func(n *dagger.Node) bool {
		for _, filter := range input.Filter {
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
		Node: &model.Node{
			Attributes: e.Node().Raw(),
			Edges:      nil,
		},
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

func (r *queryResolver) Edges(ctx context.Context, input model.QueryEdges) ([]*model.Edge, error) {
	var edges []*model.Edge
	dagger.RangeEdgeTypes(dagger.StringType(input.Type), func(edge *dagger.Edge) bool {
		for _, filter := range input.Filter {
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
		n := dagger.NewNode(edge.ID(), edge.Type(), nil)
		edges = append(edges, &model.Edge{
			Node: &model.Node{
				Attributes: n.Raw(),
				Edges:      nil,
			},
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
