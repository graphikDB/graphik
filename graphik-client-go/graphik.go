package graphik

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/gen/go"
	"github.com/golang/protobuf/ptypes/empty"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Options struct {
	retry       uint
	tokenSource oauth2.TokenSource
}

type Opt func(o *Options)

func WithRetry(retry uint) Opt {
	return func(o *Options) {
		o.retry = retry
	}
}

func WithTokenSource(tokenSource oauth2.TokenSource) Opt {
	return func(o *Options) {
		o.tokenSource = tokenSource
	}
}

func unaryAuth(tokenSource oauth2.TokenSource) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, err := toContext(ctx, tokenSource)
		if err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func streamAuth(tokenSource oauth2.TokenSource) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, err := toContext(ctx, tokenSource)
		if err != nil {
			return nil, err
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func NewClient(ctx context.Context, target string, opts ...Opt) (*Client, error) {
	dialopts := []grpc.DialOption{grpc.WithInsecure()}
	var uinterceptors []grpc.UnaryClientInterceptor
	var sinterceptors []grpc.StreamClientInterceptor
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	if options.tokenSource != nil {
		uinterceptors = append(uinterceptors, unaryAuth(options.tokenSource))
		sinterceptors = append(sinterceptors, streamAuth(options.tokenSource))
	}
	if options.retry > 0 {
		uinterceptors = append(uinterceptors, grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(options.retry),
		))
		sinterceptors = append(sinterceptors, grpc_retry.StreamClientInterceptor(
			grpc_retry.WithMax(options.retry),
		))
	}
	dialopts = append(dialopts,
		grpc.WithChainUnaryInterceptor(uinterceptors...),
		grpc.WithChainStreamInterceptor(sinterceptors...),
	)
	conn, err := grpc.DialContext(ctx, target, dialopts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		graph: apipb.NewDatabaseServiceClient(conn),
	}, nil
}

type Client struct {
	graph apipb.DatabaseServiceClient
}

func toContext(ctx context.Context, tokenSource oauth2.TokenSource) (context.Context, error) {
	token, err := tokenSource.Token()
	if err != nil {
		return ctx, err
	}
	return metadata.AppendToOutgoingContext(
		ctx,
		"Authorization", fmt.Sprintf("Bearer %v", token.AccessToken),
	), nil
}

func (c *Client) Me(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.Me(ctx, in, opts...)
}

func (c *Client) CreateDoc(ctx context.Context, in *apipb.DocConstructor, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.CreateDoc(ctx, in, opts...)
}

func (c *Client) CreateDocs(ctx context.Context, in *apipb.DocConstructors, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.CreateDocs(ctx, in, opts...)
}

func (c *Client) GetDoc(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.GetDoc(ctx, in, opts...)
}

func (c *Client) SearchDocs(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.SearchDocs(ctx, in, opts...)
}

func (c *Client) EditDoc(ctx context.Context, in *apipb.Edit, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.EditDoc(ctx, in, opts...)
}

func (c *Client) EditDocs(ctx context.Context, in *apipb.EFilter, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.EditDocs(ctx, in, opts...)
}

func (c *Client) CreateConnection(ctx context.Context, in *apipb.ConnectionConstructor, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.CreateConnection(ctx, in, opts...)
}

func (c *Client) CreateConnections(ctx context.Context, in *apipb.ConnectionConstructors, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.CreateConnections(ctx, in, opts...)
}

func (c *Client) GetConnection(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.GetConnection(ctx, in, opts...)
}

func (c *Client) SearchConnections(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.SearchConnections(ctx, in, opts...)
}

func (c *Client) EditConnection(ctx context.Context, in *apipb.Edit, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.EditConnection(ctx, in, opts...)
}

func (c *Client) EditConnections(ctx context.Context, in *apipb.EFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.EditConnections(ctx, in, opts...)
}

func (c *Client) ConnectionsFrom(ctx context.Context, in *apipb.CFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsFrom(ctx, in, opts...)
}

func (c *Client) ConnectionsTo(ctx context.Context, in *apipb.CFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsTo(ctx, in, opts...)
}

func (c *Client) Publish(ctx context.Context, in *apipb.OutboundMessage, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Publish(ctx, in, opts...)
}

func (c *Client) Subscribe(ctx context.Context, in *apipb.ChanFilter, handler func(msg *apipb.Message) bool, opts ...grpc.CallOption) error {
	stream, err := c.graph.Subscribe(ctx, in, opts...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := stream.Recv()
			if err != nil {
				return err
			}
			if !handler(msg) {
				return nil
			}
		}
	}
}

func (c *Client) SubscribeChanges(ctx context.Context, in *apipb.ExprFilter, handler func(change *apipb.Change) bool, opts ...grpc.CallOption) error {
	stream, err := c.graph.SubscribeChanges(ctx, in, opts...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			change, err := stream.Recv()
			if err != nil {
				return err
			}
			if !handler(change) {
				return nil
			}
		}
	}
}

func (c *Client) PushDocConstructors(ctx context.Context, ch <-chan *apipb.DocConstructor, opts ...grpc.CallOption) error {
	stream, err := c.graph.PushDocConstructors(ctx, opts...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (c *Client) PushConnectionConstructors(ctx context.Context, ch <-chan *apipb.ConnectionConstructor, opts ...grpc.CallOption) error {
	stream, err := c.graph.PushConnectionConstructors(ctx, opts...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (c *Client) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Pong, error) {
	return c.graph.Ping(ctx, in, opts...)
}

func (c *Client) GetSchema(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Schema, error) {
	return c.graph.GetSchema(ctx, in, opts...)
}

func (c *Client) SetIndexes(ctx context.Context, in *apipb.Indexes, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.SetIndexes(ctx, in, opts...)
}

func (c *Client) SetAuthorizers(ctx context.Context, in *apipb.Authorizers, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.SetAuthorizers(ctx, in, opts...)
}

func (c *Client) SetTypeValidators(ctx context.Context, in *apipb.TypeValidators, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.SetTypeValidators(ctx, in, opts...)
}
