package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	apipb "github.com/autom8ter/graphik/gen/go"
	generated1 "github.com/autom8ter/graphik/gen/gql/generated"
	"github.com/autom8ter/graphik/gen/gql/model"
	"github.com/autom8ter/graphik/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *mutationResolver) CreateDoc(ctx context.Context, input model.DocConstructor) (*model.Doc, error) {
	doc, err := r.client.CreateDoc(ctx, protoDocC(&input))
	if err != nil {
		return nil, err
	}
	return gqlDoc(doc), nil
}

func (r *mutationResolver) EditDoc(ctx context.Context, input model.Edit) (*model.Doc, error) {
	res, err := r.client.EditDoc(ctx, protoEdit(&input))
	if err != nil {
		return nil, err
	}
	return gqlDoc(res), nil
}

func (r *mutationResolver) EditDocs(ctx context.Context, input model.EFilter) (*model.Docs, error) {
	docs, err := r.client.EditDocs(ctx, protoEFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlDocs(docs), nil
}

func (r *mutationResolver) DelDoc(ctx context.Context, input model.RefInput) (*emptypb.Empty, error) {
	return r.client.DelDoc(ctx, protoIRef(&input))
}

func (r *mutationResolver) DelDocs(ctx context.Context, input model.Filter) (*emptypb.Empty, error) {
	return r.client.DelDocs(ctx, protoFilter(&input))
}

func (r *mutationResolver) CreateConnection(ctx context.Context, input model.ConnectionConstructor) (*model.Connection, error) {
	res, err := r.client.CreateConnection(ctx, protoConnectionC(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnection(res), nil
}

func (r *mutationResolver) EditConnection(ctx context.Context, input model.Edit) (*model.Connection, error) {
	res, err := r.client.EditConnection(ctx, protoEdit(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnection(res), nil
}

func (r *mutationResolver) EditConnections(ctx context.Context, input model.EFilter) (*model.Connections, error) {
	connections, err := r.client.EditConnections(ctx, protoEFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnections(connections), nil
}

func (r *mutationResolver) DelConnection(ctx context.Context, input model.RefInput) (*emptypb.Empty, error) {
	return r.client.DelConnection(ctx, protoIRef(&input))
}

func (r *mutationResolver) DelConnections(ctx context.Context, input model.Filter) (*emptypb.Empty, error) {
	return r.client.DelConnections(ctx, protoFilter(&input))
}

func (r *mutationResolver) Publish(ctx context.Context, input model.OutboundMessage) (*emptypb.Empty, error) {
	return r.client.Publish(ctx, &apipb.OutboundMessage{
		Channel: input.Channel,
		Data:    apipb.NewStruct(input.Data),
	})
}

func (r *mutationResolver) SetIndexes(ctx context.Context, input model.IndexesInput) (*emptypb.Empty, error) {
	var indexes []*apipb.Index
	for _, index := range input.Indexes {
		indexes = append(indexes, protoIndex(index))
	}
	return r.client.SetIndexes(ctx, &apipb.Indexes{
		Indexes: indexes,
	})
}

func (r *mutationResolver) SetAuthorizers(ctx context.Context, input model.AuthorizersInput) (*emptypb.Empty, error) {
	var authorizers []*apipb.Authorizer
	for _, auth := range input.Authorizers {
		authorizers = append(authorizers, protoAuthorizer(auth))
	}
	return r.client.SetAuthorizers(ctx, &apipb.Authorizers{
		Authorizers: authorizers,
	})
}

func (r *mutationResolver) SetTypeValidators(ctx context.Context, input model.TypeValidatorsInput) (*emptypb.Empty, error) {
	var validators []*apipb.TypeValidator
	for _, validator := range input.Validators {
		validators = append(validators, protoTypeValidator(validator))
	}
	return r.client.SetTypeValidators(ctx, &apipb.TypeValidators{
		Validators: validators,
	})
}

func (r *queryResolver) Ping(ctx context.Context, input *emptypb.Empty) (*model.Pong, error) {
	res, err := r.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return &model.Pong{Message: res.GetMessage()}, nil
}

func (r *queryResolver) GetSchema(ctx context.Context, input *emptypb.Empty) (*model.Schema, error) {
	res, err := r.client.GetSchema(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return gqlSchema(res), nil
}

func (r *queryResolver) Me(ctx context.Context, input *emptypb.Empty) (*model.Doc, error) {
	res, err := r.client.Me(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return gqlDoc(res), nil
}

func (r *queryResolver) GetDoc(ctx context.Context, input model.RefInput) (*model.Doc, error) {
	res, err := r.client.GetDoc(ctx, protoIRef(&input))
	if err != nil {
		return nil, err
	}
	return gqlDoc(res), nil
}

func (r *queryResolver) SearchDocs(ctx context.Context, input model.Filter) (*model.Docs, error) {
	res, err := r.client.SearchDocs(ctx, protoFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlDocs(res), nil
}

func (r *queryResolver) Traverse(ctx context.Context, input model.TFilter) (*model.Traversals, error) {
	res, err := r.client.Traverse(ctx, protoDepthFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlTraversals(res), nil
}

func (r *queryResolver) GetConnection(ctx context.Context, input model.RefInput) (*model.Connection, error) {
	res, err := r.client.GetConnection(ctx, protoIRef(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnection(res), nil
}

func (r *queryResolver) SearchConnections(ctx context.Context, input model.Filter) (*model.Connections, error) {
	res, err := r.client.SearchConnections(ctx, protoFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) ConnectionsFrom(ctx context.Context, input model.CFilter) (*model.Connections, error) {
	res, err := r.client.ConnectionsFrom(ctx, protoConnectionFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) ConnectionsTo(ctx context.Context, input model.CFilter) (*model.Connections, error) {
	res, err := r.client.ConnectionsFrom(ctx, protoConnectionFilter(&input))
	if err != nil {
		return nil, err
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) AggregateDocs(ctx context.Context, input model.AggFilter) (interface{}, error) {
	res, err := r.client.AggregateDocs(ctx, protoAggFilter(&input))
	if err != nil {
		return nil, err
	}
	return res.GetNumberValue(), nil
}

func (r *queryResolver) AggregateConnections(ctx context.Context, input model.AggFilter) (interface{}, error) {
	res, err := r.client.AggregateConnections(ctx, protoAggFilter(&input))
	if err != nil {
		return nil, err
	}
	return res.GetNumberValue(), nil
}

func (r *queryResolver) SearchAndConnect(ctx context.Context, input model.SConnectFilter) (*model.Connections, error) {
	connections, err := r.client.SearchAndConnect(ctx, &apipb.SConnectFilter{
		Filter:     protoFilter(input.Filter),
		Gtype:      input.Gtype,
		Attributes: apipb.NewStruct(input.Attributes),
		Directed:   input.Directed,
		From:       protoIRef(input.From),
	})
	if err != nil {
		return nil, err
	}
	return gqlConnections(connections), nil
}

func (r *subscriptionResolver) Subscribe(ctx context.Context, input model.ChanFilter) (<-chan *model.Message, error) {
	ch := make(chan *model.Message)
	stream, err := r.client.Subscribe(ctx, protoChanFilter(&input))
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
				ch <- &model.Message{
					Channel:   msg.GetChannel(),
					Data:      msg.GetData().AsMap(),
					Sender:    gqlRef(msg.GetSender()),
					Timestamp: msg.GetTimestamp().AsTime(),
				}
			}
		}
	}()
	return ch, nil
}

// Mutation returns generated1.MutationResolver implementation.
func (r *Resolver) Mutation() generated1.MutationResolver { return &mutationResolver{r} }

// Query returns generated1.QueryResolver implementation.
func (r *Resolver) Query() generated1.QueryResolver { return &queryResolver{r} }

// Subscription returns generated1.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated1.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
