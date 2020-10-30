package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

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
	return r.store.Node(ctx, input)
}

func (r *queryResolver) Nodes(ctx context.Context, input model.Filter) ([]*model.Node, error) {
	return r.store.Nodes(ctx, input)
}

func (r *queryResolver) SearchNodes(ctx context.Context, input model.Search) (*model.Results, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Edge(ctx context.Context, input model.ForeignKey) (*model.Edge, error) {
	return r.store.Edge(ctx, input)
}

func (r *queryResolver) Edges(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	return r.store.Edges(ctx, input)
}

func (r *queryResolver) SearchEdges(ctx context.Context, input model.Search) (*model.Results, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
