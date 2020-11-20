package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/gql/generated"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *mutationResolver) CreateNode(ctx context.Context, input apipb.NodeConstructor) (*apipb.Node, error) {

	return r.client.CreateNode(ctx, &input)
}

func (r *mutationResolver) PatchNode(ctx context.Context, input apipb.Patch) (*apipb.Node, error) {
	return r.client.PatchNode(ctx, &input)
}

func (r *mutationResolver) DelNode(ctx context.Context, input apipb.Path) (*emptypb.Empty, error) {
	return r.client.DelNode(ctx, &input)
}

func (r *mutationResolver) CreateEdge(ctx context.Context, input apipb.EdgeConstructor) (*apipb.Edge, error) {
	return r.client.CreateEdge(ctx, &input)
}

func (r *mutationResolver) PatchEdge(ctx context.Context, input apipb.Patch) (*apipb.Edge, error) {
	return r.client.PatchEdge(ctx, &input)
}

func (r *mutationResolver) DelEdge(ctx context.Context, input apipb.Path) (*emptypb.Empty, error) {
	return r.client.DelEdge(ctx, &input)
}

func (r *nodeDetailResolver) EdgesFrom(ctx context.Context, obj *apipb.NodeDetail) ([]*apipb.EdgeDetail, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *nodeDetailResolver) EdgesTo(ctx context.Context, obj *apipb.NodeDetail) ([]*apipb.EdgeDetail, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Ping(ctx context.Context, input *emptypb.Empty) (*apipb.Pong, error) {
	return r.client.Ping(ctx, &emptypb.Empty{})
}

func (r *queryResolver) GetSchema(ctx context.Context, input *emptypb.Empty) (*apipb.Schema, error) {
	return r.client.GetSchema(ctx, &emptypb.Empty{})
}

func (r *queryResolver) GetNode(ctx context.Context, input apipb.Path) (*apipb.Node, error) {
	return r.client.GetNode(ctx, &input)
}

func (r *queryResolver) SearchNodes(ctx context.Context, input apipb.Filter) (*apipb.Nodes, error) {
	return r.client.SearchNodes(ctx, &input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// NodeDetail returns generated.NodeDetailResolver implementation.
func (r *Resolver) NodeDetail() generated.NodeDetailResolver { return &nodeDetailResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type nodeDetailResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
