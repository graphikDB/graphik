package gql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	generated1 "github.com/graphikDB/graphik/gen/gql/go/generated"
	"github.com/graphikDB/graphik/gen/gql/go/model"
	apipb "github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (r *mutationResolver) CreateDoc(ctx context.Context, input model.DocConstructor) (*model.Doc, error) {
	doc, err := r.client.CreateDoc(ctx, protoDocC(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDoc(doc), nil
}

func (r *mutationResolver) CreateDocs(ctx context.Context, input model.DocConstructors) (*model.Docs, error) {
	docs, err := r.client.CreateDocs(ctx, protoDocCs(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDocs(docs), nil
}

func (r *mutationResolver) PutDoc(ctx context.Context, input *model.PutDoc) (*model.Doc, error) {
	res, err := r.client.PutDoc(ctx, &apipb.Doc{
		Ref: &apipb.Ref{
			Gtype: input.Ref.Gtype,
			Gid:   input.Ref.Gid,
		},
		Attributes: apipb.NewStruct(input.Attributes),
	})
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDoc(res), nil
}

func (r *mutationResolver) PutDocs(ctx context.Context, input *model.PutDocs) (*model.Docs, error) {
	var docs = &apipb.Docs{}
	var err error
	for _, d := range input.Docs {
		docs.Docs = append(docs.Docs, &apipb.Doc{
			Ref: &apipb.Ref{
				Gtype: d.Ref.Gtype,
				Gid:   d.Ref.Gid,
			},
			Attributes: apipb.NewStruct(d.Attributes),
		})
	}
	docs, err = r.client.PutDocs(ctx, docs)
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDocs(docs), nil
}

func (r *mutationResolver) EditDoc(ctx context.Context, input model.Edit) (*model.Doc, error) {
	res, err := r.client.EditDoc(ctx, protoEdit(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDoc(res), nil
}

func (r *mutationResolver) EditDocs(ctx context.Context, input model.EditFilter) (*model.Docs, error) {
	docs, err := r.client.EditDocs(ctx, protoEditFilter(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDocs(docs), nil
}

func (r *mutationResolver) DelDoc(ctx context.Context, input model.RefInput) (*emptypb.Empty, error) {
	if e, err := r.client.DelDoc(ctx, protoIRef(input)); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) DelDocs(ctx context.Context, input model.Filter) (*emptypb.Empty, error) {
	if e, err := r.client.DelDocs(ctx, protoFilter(input)); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) CreateConnection(ctx context.Context, input model.ConnectionConstructor) (*model.Connection, error) {
	res, err := r.client.CreateConnection(ctx, protoConnectionC(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnection(res), nil
}

func (r *mutationResolver) CreateConnections(ctx context.Context, input model.ConnectionConstructors) (*model.Connections, error) {
	connections, err := r.client.CreateConnections(ctx, protoConnectionCs(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(connections), nil
}

func (r *mutationResolver) PutConnection(ctx context.Context, input *model.PutConnection) (*model.Connection, error) {
	res, err := r.client.PutConnection(ctx, protoConnection(&model.Connection{
		Ref: &model.Ref{
			Gtype: input.Ref.Gtype,
			Gid:   input.Ref.Gid,
		},
		Attributes: input.Attributes,
		Directed:   input.Directed,
		From: &model.Ref{
			Gtype: input.From.Gtype,
			Gid:   input.From.Gid,
		},
		To: &model.Ref{
			Gtype: input.To.Gtype,
			Gid:   input.To.Gid,
		},
	}))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnection(res), nil
}

func (r *mutationResolver) PutConnections(ctx context.Context, input *model.PutConnections) (*model.Connections, error) {
	var connections = &apipb.Connections{}
	var err error
	for _, d := range input.Connections {
		connections.Connections = append(connections.Connections, &apipb.Connection{
			Ref: &apipb.Ref{
				Gtype: d.Ref.Gtype,
				Gid:   d.Ref.Gid,
			},
			Attributes: apipb.NewStruct(d.Attributes),
			Directed:   d.Directed,
			From: &apipb.Ref{
				Gtype: d.From.Gtype,
				Gid:   d.From.Gid,
			},
			To: &apipb.Ref{
				Gtype: d.To.Gtype,
				Gid:   d.To.Gid,
			},
		})
	}
	connections, err = r.client.PutConnections(ctx, connections)
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(connections), nil
}

func (r *mutationResolver) EditConnection(ctx context.Context, input model.Edit) (*model.Connection, error) {
	res, err := r.client.EditConnection(ctx, protoEdit(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnection(res), nil
}

func (r *mutationResolver) EditConnections(ctx context.Context, input model.EditFilter) (*model.Connections, error) {
	connections, err := r.client.EditConnections(ctx, protoEditFilter(input))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(connections), nil
}

func (r *mutationResolver) DelConnection(ctx context.Context, input model.RefInput) (*emptypb.Empty, error) {
	if e, err := r.client.DelConnection(ctx, protoIRef(input)); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) DelConnections(ctx context.Context, input model.Filter) (*emptypb.Empty, error) {
	if e, err := r.client.DelConnections(ctx, protoFilter(input)); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) Broadcast(ctx context.Context, input model.OutboundMessage) (*emptypb.Empty, error) {
	if e, err := r.client.Broadcast(ctx, &apipb.OutboundMessage{
		Channel: input.Channel,
		Data:    apipb.NewStruct(input.Data),
	}); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) SetIndexes(ctx context.Context, input model.IndexesInput) (*emptypb.Empty, error) {
	var indexes []*apipb.Index
	for _, index := range input.Indexes {
		indexes = append(indexes, protoIndex(index))
	}
	if e, err := r.client.SetIndexes(ctx, &apipb.Indexes{
		Indexes: indexes,
	}); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) SetAuthorizers(ctx context.Context, input model.AuthorizersInput) (*emptypb.Empty, error) {
	var authorizers []*apipb.Authorizer
	for _, auth := range input.Authorizers {
		authorizers = append(authorizers, protoAuthorizer(auth))
	}
	if e, err := r.client.SetAuthorizers(ctx, &apipb.Authorizers{
		Authorizers: authorizers,
	}); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) SetConstraints(ctx context.Context, input model.ConstraintsInput) (*emptypb.Empty, error) {
	var validators []*apipb.Constraint
	for _, validator := range input.Constraints {
		validators = append(validators, protoConstraint(validator))
	}
	if e, err := r.client.SetConstraints(ctx, &apipb.Constraints{
		Constraints: validators,
	}); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) SetTriggers(ctx context.Context, input model.TriggersInput) (*emptypb.Empty, error) {
	var triggers []*apipb.Trigger
	for _, trigger := range input.Triggers {
		triggers = append(triggers, protoTrigger(trigger))
	}
	if e, err := r.client.SetTriggers(ctx, &apipb.Triggers{
		Triggers: triggers,
	}); err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	} else {
		return e, nil
	}
}

func (r *mutationResolver) SearchAndConnect(ctx context.Context, input model.SearchConnectFilter) (*model.Connections, error) {
	connections, err := r.client.SearchAndConnect(ctx, &apipb.SearchConnectFilter{
		Filter:     protoFilter(*input.Filter),
		Gtype:      input.Gtype,
		Attributes: apipb.NewStruct(input.Attributes),
		Directed:   input.Directed,
		From:       protoIRef(*input.From),
	})
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(connections), nil
}

func (r *mutationResolver) SearchAndConnectMe(ctx context.Context, input model.SearchConnectMeFilter) (*model.Connections, error) {
	connections, err := r.client.SearchAndConnectMe(ctx, &apipb.SearchConnectMeFilter{
		Filter:     protoFilter(*input.Filter),
		Gtype:      input.Gtype,
		Attributes: apipb.NewStruct(input.Attributes),
		Directed:   input.Directed,
	})
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(connections), nil
}

func (r *queryResolver) GetSchema(ctx context.Context, where *emptypb.Empty) (*model.Schema, error) {
	res, err := r.client.GetSchema(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlSchema(res), nil
}

func (r *queryResolver) Me(ctx context.Context, where *emptypb.Empty) (*model.Doc, error) {
	res, err := r.client.Me(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDoc(res), nil
}

func (r *queryResolver) GetDoc(ctx context.Context, where model.RefInput) (*model.Doc, error) {
	res, err := r.client.GetDoc(ctx, protoIRef(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDoc(res), nil
}

func (r *queryResolver) SearchDocs(ctx context.Context, where model.Filter) (*model.Docs, error) {
	res, err := r.client.SearchDocs(ctx, protoFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlDocs(res), nil
}

func (r *queryResolver) Traverse(ctx context.Context, where model.TraverseFilter) (*model.Traversals, error) {
	res, err := r.client.Traverse(ctx, protoTraverseFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}

	return gqlTraversals(res), nil
}

func (r *queryResolver) TraverseMe(ctx context.Context, where model.TraverseMeFilter) (*model.Traversals, error) {
	res, err := r.client.TraverseMe(ctx, protoTraverseMeFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlTraversals(res), nil
}

func (r *queryResolver) GetConnection(ctx context.Context, where model.RefInput) (*model.Connection, error) {
	res, err := r.client.GetConnection(ctx, protoIRef(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnection(res), nil
}

func (r *queryResolver) ExistsDoc(ctx context.Context, where model.ExistsFilter) (bool, error) {
	res, err := r.client.ExistsDoc(ctx, protoExists(where))
	if err != nil {
		return false, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *queryResolver) ExistsConnection(ctx context.Context, where model.ExistsFilter) (bool, error) {
	res, err := r.client.ExistsConnection(ctx, protoExists(where))
	if err != nil {
		return false, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *queryResolver) HasDoc(ctx context.Context, where model.RefInput) (bool, error) {
	res, err := r.client.HasDoc(ctx, protoIRef(where))
	if err != nil {
		return false, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *queryResolver) HasConnection(ctx context.Context, where model.RefInput) (bool, error) {
	res, err := r.client.HasConnection(ctx, protoIRef(where))
	if err != nil {
		return false, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *queryResolver) SearchConnections(ctx context.Context, where model.Filter) (*model.Connections, error) {
	res, err := r.client.SearchConnections(ctx, protoFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) ConnectionsFrom(ctx context.Context, where model.ConnectFilter) (*model.Connections, error) {
	res, err := r.client.ConnectionsFrom(ctx, protoConnectionFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) ConnectionsTo(ctx context.Context, where model.ConnectFilter) (*model.Connections, error) {
	res, err := r.client.ConnectionsFrom(ctx, protoConnectionFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return gqlConnections(res), nil
}

func (r *queryResolver) AggregateDocs(ctx context.Context, where model.AggFilter) (float64, error) {
	res, err := r.client.AggregateDocs(ctx, protoAggFilter(where))
	if err != nil {
		return 0, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *queryResolver) AggregateConnections(ctx context.Context, where model.AggFilter) (float64, error) {
	res, err := r.client.AggregateConnections(ctx, protoAggFilter(where))
	if err != nil {
		return 0, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
	}
	return res.GetValue(), nil
}

func (r *subscriptionResolver) Stream(ctx context.Context, where model.StreamFilter) (<-chan *model.Message, error) {
	ch := make(chan *model.Message)
	stream, err := r.client.Stream(ctx, protoStreamFilter(where))
	if err != nil {
		return nil, &gqlerror.Error{
			Message: err.Error(),
			Path:    graphql.GetPath(ctx),
			Extensions: map[string]interface{}{
				"code": status.Code(err).String(),
			},
		}
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
					r.logger.Error("failed to receive subsription message", zap.Error(err))
					continue
				}
				ch <- &model.Message{
					Channel:   msg.GetChannel(),
					Data:      msg.GetData().AsMap(),
					User:      gqlRef(msg.GetUser()),
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
