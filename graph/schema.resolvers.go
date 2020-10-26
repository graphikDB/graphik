package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/autom8ter/dagger/primitive"

	"github.com/autom8ter/dagger"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/graph/model"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input model.NewNode) (*model.Node, error) {
	node := dagger.NewNode(input.Type, "", input.Attributes)
	return &model.Node{
		ID:         node.ID(),
		Type:       node.Type(),
		Attributes: node.Raw(),
	}, nil
}

func (r *queryResolver) Nodes(ctx context.Context, input model.QueryNodes) ([]*model.Node, error) {
	var nodes []*model.Node
	dagger.RangeNodeTypes(primitive.StringType(input.Type), func(n *dagger.Node) bool {
		nodes = append(nodes, &model.Node{
			ID:         n.ID(),
			Type:       n.Type(),
			Attributes: n.Raw(),
		})
		return true
	})
	return nodes, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
