package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	apipb "github.com/autom8ter/graphik/gen/go/api"
	"github.com/autom8ter/graphik/gql/generated"
	"github.com/autom8ter/graphik/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *metadataResolver) Sequence(ctx context.Context, obj *apipb.Metadata) (int, error) {
	return int(obj.Sequence), nil
}

func (r *metadataResolver) Version(ctx context.Context, obj *apipb.Metadata) (int, error) {
	return int(obj.Version), nil
}

func (r *mutationResolver) CreateDoc(ctx context.Context, input apipb.DocConstructor) (*apipb.Doc, error) {
	return r.client.CreateDoc(ctx, &input)
}

func (r *mutationResolver) PatchDoc(ctx context.Context, input apipb.Patch) (*apipb.Doc, error) {
	return r.client.PatchDoc(ctx, &input)
}

func (r *mutationResolver) PatchDocs(ctx context.Context, input apipb.PatchFilter) (*apipb.Docs, error) {
	return r.client.PatchDocs(ctx, &input)
}

func (r *mutationResolver) CreateConnection(ctx context.Context, input apipb.ConnectionConstructor) (*apipb.Connection, error) {
	return r.client.CreateConnection(ctx, &input)
}

func (r *mutationResolver) PatchConnection(ctx context.Context, input apipb.Patch) (*apipb.Connection, error) {
	return r.client.PatchConnection(ctx, &input)
}

func (r *mutationResolver) PatchConnections(ctx context.Context, input apipb.PatchFilter) (*apipb.Connections, error) {
	return r.client.PatchConnections(ctx, &input)
}

func (r *mutationResolver) Publish(ctx context.Context, input apipb.OutboundMessage) (*emptypb.Empty, error) {
	return r.client.Publish(ctx, &input)
}

func (r *queryResolver) Ping(ctx context.Context, input *emptypb.Empty) (*apipb.Pong, error) {
	return r.client.Ping(ctx, &emptypb.Empty{})
}

func (r *queryResolver) GetSchema(ctx context.Context, input *emptypb.Empty) (*apipb.Schema, error) {
	return r.client.GetSchema(ctx, &emptypb.Empty{})
}

func (r *queryResolver) Me(ctx context.Context, input *apipb.MeFilter) (*apipb.DocDetail, error) {
	return r.client.Me(ctx, input)
}

func (r *queryResolver) GetDoc(ctx context.Context, input apipb.Path) (*apipb.Doc, error) {
	return r.client.GetDoc(ctx, &input)
}

func (r *queryResolver) SearchDocs(ctx context.Context, input apipb.Filter) (*apipb.Docs, error) {
	return r.client.SearchDocs(ctx, &input)
}

func (r *queryResolver) GetConnection(ctx context.Context, input apipb.Path) (*apipb.Connection, error) {
	return r.client.GetConnection(ctx, &input)
}

func (r *queryResolver) SearchConnections(ctx context.Context, input apipb.Filter) (*apipb.Connections, error) {
	return r.client.SearchConnections(ctx, &input)
}

func (r *queryResolver) ConnectionsFrom(ctx context.Context, input apipb.ConnectionFilter) (*apipb.Connections, error) {
	return r.client.ConnectionsFrom(ctx, &input)
}

func (r *queryResolver) ConnectionsTo(ctx context.Context, input apipb.ConnectionFilter) (*apipb.Connections, error) {
	return r.client.ConnectionsFrom(ctx, &input)
}

func (r *subscriptionResolver) Subscribe(ctx context.Context, input apipb.ChannelFilter) (<-chan *apipb.Message, error) {
	ch := make(chan *apipb.Message)
	stream, err := r.client.Subscribe(ctx, &input)
	if err != nil {
		return nil, err
	}
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
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

func (r *subscriptionResolver) SubscribeChanges(ctx context.Context, input apipb.ExpressionFilter) (<-chan *apipb.Change, error) {
	ch := make(chan *apipb.Change)
	stream, err := r.client.SubscribeChanges(ctx, &input)
	if err != nil {
		return nil, err
	}
	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			default:
				msg, err := stream.Recv()
				if err != nil {
					logger.Error("failed to receive change", zap.Error(err))
					continue
				}
				ch <- msg
			}
		}
	}()
	return ch, nil
}

// Metadata returns generated.MetadataResolver implementation.
func (r *Resolver) Metadata() generated.MetadataResolver { return &metadataResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type metadataResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
