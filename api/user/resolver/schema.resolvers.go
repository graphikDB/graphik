package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/autom8ter/graphik/api/model"
	"github.com/autom8ter/graphik/api/user/generated"
)

func (r *queryResolver) Me(ctx context.Context, input *model.Empty) (*model.Node, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("nonexistant node")
	}
	return n, nil
}

func (r *queryResolver) EdgesFrom(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("nonexistant node")
	}
	return r.runtime.EdgesFrom(ctx, n.Path, input)
}

func (r *queryResolver) EdgesTo(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("nonexistant node")
	}
	return r.runtime.EdgesTo(ctx, n.Path, input)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
