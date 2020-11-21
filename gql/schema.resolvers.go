package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/gql/generated"
	"github.com/autom8ter/graphik/logger"
	"go.uber.org/zap"
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

func (r *mutationResolver) Publish(ctx context.Context, input *apipb.OutboundMessage) (*emptypb.Empty, error) {
	return r.client.Publish(ctx, input)
}

func (r *queryResolver) Ping(ctx context.Context, input *emptypb.Empty) (*apipb.Pong, error) {
	return r.client.Ping(ctx, &emptypb.Empty{})
}

func (r *queryResolver) GetSchema(ctx context.Context, input *emptypb.Empty) (*apipb.Schema, error) {
	return r.client.GetSchema(ctx, &emptypb.Empty{})
}

func (r *queryResolver) Me(ctx context.Context, input *apipb.MeFilter) (*apipb.NodeDetail, error) {
	return r.client.Me(ctx, input)
}

func (r *queryResolver) GetNode(ctx context.Context, input apipb.Path) (*apipb.Node, error) {
	return r.client.GetNode(ctx, &input)
}

func (r *queryResolver) SearchNodes(ctx context.Context, input apipb.Filter) (*apipb.Nodes, error) {
	return r.client.SearchNodes(ctx, &input)
}

func (r *queryResolver) GetEdge(ctx context.Context, input apipb.Path) (*apipb.Edge, error) {
	return r.client.GetEdge(ctx, &input)
}

func (r *queryResolver) SearchEdges(ctx context.Context, input apipb.Filter) (*apipb.Edges, error) {
	return r.client.SearchEdges(ctx, &input)
}

func (r *queryResolver) EdgesFrom(ctx context.Context, input apipb.EdgeFilter) (*apipb.Edges, error) {
	return r.client.EdgesFrom(ctx, &input)
}

func (r *queryResolver) EdgesTo(ctx context.Context, input apipb.EdgeFilter) (*apipb.Edges, error) {
	return r.client.EdgesFrom(ctx, &input)
}

func (r *subscriptionResolver) Subscribe(ctx context.Context, input apipb.ChannelFilter) (<-chan *apipb.Message, error) {
	ch := make(chan *apipb.Message)
	stream, err := r.client.Subscribe(ctx, &input)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				msg, err := stream.Recv()
				if err != nil {
					logger.Error("failed to receive subsription message", zap.Error(err))
					continue
				}
				ch <- msg
			}
		}
	}()
	return ch, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
