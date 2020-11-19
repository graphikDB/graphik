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
			ctx, err := toContext(ctx, tokenSource)
			if err != nil {
				return err
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
		grpc.WithChainStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			ctx, err := toContext(ctx, tokenSource)
			if err != nil {
				return nil, err
			}
			return streamer(ctx, desc, cc, method, opts...)
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		graph:       apipb.NewGraphServiceClient(conn),
		tokenSource: tokenSource,
	}, nil
}

type Client struct {
	graph       apipb.GraphServiceClient
	tokenSource oauth2.TokenSource
}

func toContext(ctx context.Context, tokenSource oauth2.TokenSource) (context.Context, error) {
	token, err := tokenSource.Token()
	if err != nil {
		return ctx, err
	}
	id := token.Extra("id_token")
	ctx = metadata.AppendToOutgoingContext(context.Background(), "Authorization", fmt.Sprintf("Bearer %v", id))
	return ctx, nil
}

func (c *Client) Me(ctx context.Context, in *apipb.MeFilter, opts ...grpc.CallOption) (*apipb.NodeDetail, error) {
	return c.graph.Me(ctx, in, opts...)
}

func (c *Client) CreateNode(ctx context.Context, in *apipb.NodeConstructor, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.CreateNode(ctx, in, opts...)
}

func (c *Client) CreateNodes(ctx context.Context, in *apipb.NodeConstructors, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.CreateNodes(ctx, in, opts...)
}

func (c *Client) GetNode(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.GetNode(ctx, in, opts...)
}

func (c *Client) SearchNodes(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.SearchNodes(ctx, in, opts...)
}

func (c *Client) PatchNode(ctx context.Context, in *apipb.Patch, opts ...grpc.CallOption) (*apipb.Node, error) {
	return c.graph.PatchNode(ctx, in, opts...)
}

func (c *Client) PatchNodes(ctx context.Context, in *apipb.Patches, opts ...grpc.CallOption) (*apipb.Nodes, error) {
	return c.graph.PatchNodes(ctx, in, opts...)
}

func (c *Client) DelNode(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.DelNode(ctx, in, opts...)
}

func (c *Client) DelNodes(ctx context.Context, in *apipb.Paths, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.DelNodes(ctx, in, opts...)
}

func (c *Client) CreateEdge(ctx context.Context, in *apipb.EdgeConstructor, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.CreateEdge(ctx, in, opts...)
}

func (c *Client) CreateEdges(ctx context.Context, in *apipb.EdgeConstructors, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.CreateEdges(ctx, in, opts...)
}

func (c *Client) GetEdge(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.GetEdge(ctx, in, opts...)
}

func (c *Client) SearchEdges(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.SearchEdges(ctx, in, opts...)
}

func (c *Client) PatchEdge(ctx context.Context, in *apipb.Patch, opts ...grpc.CallOption) (*apipb.Edge, error) {
	return c.graph.PatchEdge(ctx, in, opts...)
}

func (c *Client) PatchEdges(ctx context.Context, in *apipb.Patches, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.PatchEdges(ctx, in, opts...)
}

func (c *Client) DelEdge(ctx context.Context, in *apipb.Path, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.DelEdge(ctx, in, opts...)
}

func (c *Client) DelEdges(ctx context.Context, in *apipb.Paths, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.DelEdges(ctx, in, opts...)
}

func (c *Client) EdgesFrom(ctx context.Context, in *apipb.EdgeFilter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.EdgesFrom(ctx, in, opts...)
}

func (c *Client) EdgesTo(ctx context.Context, in *apipb.EdgeFilter, opts ...grpc.CallOption) (*apipb.Edges, error) {
	return c.graph.EdgesTo(ctx, in, opts...)
}

func (c *Client) Publish(ctx context.Context, in *apipb.OutboundMessage, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Publish(ctx, in, opts...)
}

func (c *Client) Subscribe(ctx context.Context, in *apipb.ChannelFilter, opts ...grpc.CallOption) (apipb.GraphService_SubscribeClient, error) {
	return c.graph.Subscribe(ctx, in, opts...)
}

func (c *Client) SubGraph(ctx context.Context, in *apipb.SubGraphFilter) (*apipb.Graph, error) {
	return c.graph.SubGraph(ctx, in)
}

func (c *Client) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Pong, error) {
	return c.graph.Ping(ctx, in, opts...)
}

func (c *Client) JoinCluster(ctx context.Context, in *apipb.RaftNode, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.JoinCluster(ctx, in, opts...)
}

func (c *Client) GetSchema(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Schema, error) {
	return c.graph.GetSchema(ctx, in, opts...)
}

func (c *Client) Shutdown(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Shutdown(ctx, in, opts...)
}
