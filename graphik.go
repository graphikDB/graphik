package graphik

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewClient(ctx context.Context, target string, tokenSource oauth2.TokenSource) (*Client, error) {
	conn, err := grpc.DialContext(ctx, target,
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			token, err := tokenSource.Token()
			if err != nil {
				return err
			}
			id := token.Extra("id_token")
			ctx = metadata.AppendToOutgoingContext(context.Background(), "Authorization", fmt.Sprintf("Bearer %v", id))
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	if err != nil {
		return nil, err
	}
	return &Client{
		graph:       apipb.NewGraphServiceClient(conn),
		config:      apipb.NewConfigServiceClient(conn),
		tokenSource: tokenSource,
	}, nil
}

type Client struct {
	graph       apipb.GraphServiceClient
	config      apipb.ConfigServiceClient
	tokenSource oauth2.TokenSource
}

func (c *Client) Me(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.Me(ctx, in, opts...)
}

func (c *Client) CreateNode(ctx context.Context, in *apipb.Node, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.CreateNode(ctx, in, opts...)
}

func (c *Client) CreateNodes(ctx context.Context, in *apipb.Nodes, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.CreateNodes(ctx, in, opts...)
}

func (c *Client) GetNode(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.GetNode(ctx, in, opts...)
}

func (c *Client) SearchNodes(ctx context.Context, in *apipb.TypeFilter, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.SearchNodes(ctx, in, opts...)
}

func (c *Client) PatchNode(ctx context.Context, in *apipb.Node, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.PatchNode(ctx, in, opts...)
}

func (c *Client) PatchNodes(ctx context.Context, in *apipb.Nodes, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.PatchNodes(ctx, in, opts...)
}

func (c *Client) DelNode(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Counter, error) {
	return c.graph.DelNode(ctx, in, opts...)
}

func (c *Client) DelNodes(ctx context.Context, in *apipb.Paths, opts ...grpc.CallOption) (*apipb.Counter, error) {
	return c.graph.DelNodes(ctx, in, opts...)
}

func (c *Client) CreateEdge(ctx context.Context, in *apipb.Edge, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.CreateEdge(ctx, in, opts...)
}

func (c *Client) CreateEdges(ctx context.Context, in *apipb.Edges, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.CreateEdges(ctx, in, opts...)
}

func (c *Client) GetEdge(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.GetEdge(ctx, in, opts...)
}

func (c *Client) SearchEdges(ctx context.Context, in *apipb.TypeFilter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.SearchEdges(ctx, in, opts...)
}

func (c *Client) PatchEdge(ctx context.Context, in *apipb.Edge, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.PatchEdge(ctx, in, opts...)
}

func (c *Client) PatchEdges(ctx context.Context, in *apipb.Edges, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.PatchEdges(ctx, in, opts...)
}

func (c *Client) DelEdge(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Counter, error) {
	return c.graph.DelEdge(ctx, in, opts...)
}

func (c *Client) DelEdges(ctx context.Context, in *apipb.Paths, opts ...grpc.CallOption) (*apipb.Counter, error) {
	return c.graph.DelEdges(ctx, in, opts...)
}

func (c *Client) EdgesFrom(ctx context.Context, in *apipb.EdgeFilter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.EdgesFrom(ctx, in, opts...)
}

func (c *Client) EdgesTo(ctx context.Context, in *apipb.EdgeFilter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.EdgesTo(ctx, in, opts...)
}

func (c *Client) ChangeStream(ctx context.Context, in *apipb.ChangeFilter, opts ...grpc.CallOption) (apipb.GraphService_ChangeStreamClient, error) {
	return c.graph.ChangeStream(ctx, in, opts...)
}

func (c *Client) Publish(ctx context.Context, in *apipb.OutboundMessage, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Publish(ctx, in, opts...)
}

func (c *Client) Subscribe(ctx context.Context, in *apipb.ChannelFilter, opts ...grpc.CallOption) (apipb.GraphService_SubscribeClient, error) {
	return c.graph.Subscribe(ctx, in, opts...)
}

func (c *Client) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Pong, error) {
	return c.config.Ping(ctx, in, opts...)
}

func (c *Client) JoinCluster(ctx context.Context, in *apipb.RaftNode, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.config.JoinCluster(ctx, in, opts...)
}

func (c *Client) GetAuth(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.AuthConfig, error) {
	return c.config.GetAuth(ctx, in, opts...)
}

func (c *Client) SetAuth(ctx context.Context, in *apipb.AuthConfig, opts ...grpc.CallOption) (*apipb.AuthConfig, error) {
	return c.config.SetAuth(ctx, in, opts...)
}
