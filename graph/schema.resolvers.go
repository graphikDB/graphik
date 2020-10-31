package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/autom8ter/graphik/command"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/graphik/graph/generated"
	"github.com/autom8ter/graphik/graph/model"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input model.NodeConstructor) (*model.Node, error) {
	if input.Path.ID == "" {
		random := generic.UUID()
		input.Path.ID = random
	}
	if input.Path.Type == "" {
		input.Path.Type = generic.Default
	}
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

func (r *mutationResolver) DelNode(ctx context.Context, input model.Path) (*model.Counter, error) {
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
	if input.Path.ID == "" {
		random := generic.UUID()
		input.Path.ID = random
	}
	if input.Path.Type == "" {
		input.Path.Type = generic.Default
	}
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

func (r *mutationResolver) DelEdge(ctx context.Context, input model.Path) (*model.Counter, error) {
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

func (r *queryResolver) Node(ctx context.Context, input model.Path) (*model.Node, error) {
	return r.store.Node(ctx, input)
}

func (r *queryResolver) GetNodes(ctx context.Context, input model.Filter) ([]*model.Node, error) {
	return r.store.Nodes(ctx, input)
}

func (r *queryResolver) DepthSearch(ctx context.Context, input model.DepthFilter) ([]*model.Node, error) {
	if input.Reverse != nil && *input.Reverse {
		return r.store.DepthTo(ctx, input)
	}
	return r.store.DepthFrom(ctx, input)
}

func (r *queryResolver) GetEdge(ctx context.Context, input model.Path) (*model.Edge, error) {
	return r.store.Edge(ctx, input)
}

func (r *queryResolver) GetEdges(ctx context.Context, input model.Filter) ([]*model.Edge, error) {
	return r.store.Edges(ctx, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) BreadthSearch(ctx context.Context, input model.BreadthFilter) ([]*model.Node, error) {
	panic(fmt.Errorf("not implemented"))
}
