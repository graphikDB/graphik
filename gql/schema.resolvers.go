package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	apipb "github.com/autom8ter/graphik/gen/go"
	generated1 "github.com/autom8ter/graphik/gen/gql/generated"
	"github.com/autom8ter/graphik/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *metadataResolver) Version(ctx context.Context, obj *apipb.Metadata) (int, error) {
	return int(obj.Version), nil
}

func (r *mutationResolver) CreateDoc(ctx context.Context, input apipb.DocConstructor) (*apipb.Doc, error) {
	return r.client.CreateDoc(ctx, &input)
}

func (r *mutationResolver) EditDoc(ctx context.Context, input apipb.Edit) (*apipb.Doc, error) {
	return r.client.EditDoc(ctx, &input)
}

func (r *mutationResolver) EditDocs(ctx context.Context, input apipb.EditFilter) (*apipb.Docs, error) {
	return r.client.EditDocs(ctx, &input)
}

func (r *mutationResolver) CreateConnection(ctx context.Context, input apipb.ConnectionConstructor) (*apipb.Connection, error) {
	return r.client.CreateConnection(ctx, &input)
}

func (r *mutationResolver) EditConnection(ctx context.Context, input apipb.Edit) (*apipb.Connection, error) {
	return r.client.EditConnection(ctx, &input)
}

func (r *mutationResolver) EditConnections(ctx context.Context, input apipb.EditFilter) (*apipb.Connections, error) {
	return r.client.EditConnections(ctx, &input)
}

func (r *mutationResolver) Publish(ctx context.Context, input apipb.OutboundMessage) (*emptypb.Empty, error) {
	return r.client.Publish(ctx, &input)
}

func (r *mutationResolver) SetIndexes(ctx context.Context, input apipb.Indexes) (*emptypb.Empty, error) {
	return r.client.SetIndexes(ctx, &input)
}

func (r *mutationResolver) SetAuthorizers(ctx context.Context, input apipb.Authorizers) (*emptypb.Empty, error) {
	return r.client.SetAuthorizers(ctx, &input)
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

// Metadata returns generated1.MetadataResolver implementation.
func (r *Resolver) Metadata() generated1.MetadataResolver { return &metadataResolver{r} }

// Mutation returns generated1.MutationResolver implementation.
func (r *Resolver) Mutation() generated1.MutationResolver { return &mutationResolver{r} }

// Query returns generated1.QueryResolver implementation.
func (r *Resolver) Query() generated1.QueryResolver { return &queryResolver{r} }

// Subscription returns generated1.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated1.SubscriptionResolver { return &subscriptionResolver{r} }

type metadataResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
