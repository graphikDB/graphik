package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/autom8ter/graphik/api/user/generated"
	"github.com/autom8ter/graphik/lib/generic"
	"github.com/autom8ter/graphik/lib/model"
)

func (r *mutationResolver) Patch(ctx context.Context, input map[string]interface{}) (*model.Node, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	res, err := r.runtime.Execute(&model.Command{
		Op: model.OpPatchNode,
		Value: model.Patch{
			Path:  n.Path,
			Patch: input,
		},
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Node), nil
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input model.Connection) (*model.Edge, error) {
	i := &input
	if i.Path.ID == "" {
		i.Path.ID = generic.UUID()
	}
	if i.Path.Type == "" {
		i.Path.Type = generic.Default
	}
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	res, err := r.runtime.Execute(&model.Command{
		Op: model.OpCreateEdge,
		Value: model.EdgeConstructor{
			Path:       i.Path,
			Mutual:     i.Mutual,
			Attributes: i.Attributes,
			From:       n.Path,
			To:         i.To,
		},
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
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	edge, err := r.runtime.Edge(ctx, input.Path)
	if err != nil {
		return nil, err
	}
	if edge.From != n.Path && edge.To != n.Path {
		return nil, fmt.Errorf("%s is not involved with edge %s", n.Path.String(), input.Path.String())
	}
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpPatchEdge,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Edge), nil
}

func (r *mutationResolver) DelEdge(ctx context.Context, input model.Path) (*model.Counter, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	edge, err := r.runtime.Edge(ctx, input)
	if err != nil {
		return nil, err
	}
	if edge.From != n.Path && edge.To != n.Path {
		return nil, fmt.Errorf("%s is not involved with edge %s", n.Path.String(), input.String())
	}
	res, err := r.runtime.Execute(&model.Command{
		Op:    model.OpDeleteEdge,
		Value: input,
	})
	if err != nil {
		return nil, err
	}
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res.(*model.Counter), nil
}

func (r *queryResolver) Me(ctx context.Context, input *model.Empty) (*model.Node, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	return n, nil
}

func (r *queryResolver) EdgesFrom(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	return r.runtime.EdgesFrom(ctx, n.Path, input)
}

func (r *queryResolver) EdgesTo(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	n := r.runtime.GetNode(ctx)
	if n == nil {
		return nil, errors.New("non-existant node")
	}
	return r.runtime.EdgesTo(ctx, n.Path, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
