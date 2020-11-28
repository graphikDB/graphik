package graphik

import (
	"context"
	"fmt"
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/golang/protobuf/ptypes/empty"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
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
	id := token.Extra("id_token")
	if id == nil {
		return nil, errors.New("empty id token")
	}
	return metadata.AppendToOutgoingContext(
		ctx,
		"Authorization", fmt.Sprintf("Bearer %v", token.AccessToken),
		"X-GRAPHIK-ID", id.(string),
	), nil
}

func (c *Client) Me(ctx context.Context, in *apipb.MeFilter, opts ...grpc.CallOption) (*apipb.DocDetail, error) {
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

func (c *Client) PatchDoc(ctx context.Context, in *apipb.Patch, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.PatchDoc(ctx, in, opts...)
}

func (c *Client) PatchDocs(ctx context.Context, in *apipb.PatchFilter, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.PatchDocs(ctx, in, opts...)
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

func (c *Client) PatchConnection(ctx context.Context, in *apipb.Patch, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.PatchConnection(ctx, in, opts...)
}

func (c *Client) PatchConnections(ctx context.Context, in *apipb.PatchFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.PatchConnections(ctx, in, opts...)
}

func (c *Client) ConnectionsFrom(ctx context.Context, in *apipb.ConnectionFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsFrom(ctx, in, opts...)
}

func (c *Client) ConnectionsTo(ctx context.Context, in *apipb.ConnectionFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsTo(ctx, in, opts...)
}

func (c *Client) Publish(ctx context.Context, in *apipb.OutboundMessage, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Publish(ctx, in, opts...)
}

func (c *Client) Subscribe(ctx context.Context, in *apipb.ChannelFilter, opts ...grpc.CallOption) (apipb.DatabaseService_SubscribeClient, error) {
	return c.graph.Subscribe(ctx, in, opts...)
}

func (c *Client) SubscribeChanges(ctx context.Context, in *apipb.ExpressionFilter, opts ...grpc.CallOption) (apipb.DatabaseService_SubscribeChangesClient, error) {
	return c.graph.SubscribeChanges(ctx, in, opts...)
}

func (c *Client) SubGraph(ctx context.Context, in *apipb.SubGraphFilter) (*apipb.Graph, error) {
	return c.graph.SubGraph(ctx, in)
}

func (c *Client) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Pong, error) {
	return c.graph.Ping(ctx, in, opts...)
}

func (c *Client) GetSchema(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Schema, error) {
	return c.graph.GetSchema(ctx, in, opts...)
}

func (c *Client) Shutdown(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Shutdown(ctx, in, opts...)
}
