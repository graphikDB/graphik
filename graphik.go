package graphik

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		graph: apipb.NewGraphServiceClient(conn),
	}, nil
}

type Client struct {
	graph apipb.GraphServiceClient
}

func toContext(ctx context.Context, tokenSource oauth2.TokenSource) (context.Context, error) {
	token, err := tokenSource.Token()
	if err != nil {
		return ctx, err
	}
	id := token.Extra("id_token")
	if id != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %v", id))
	} else {
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %v", token.AccessToken))
	}
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

func (c *Client) GetSchema(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Schema, error) {
	return c.graph.GetSchema(ctx, in, opts...)
}

func (c *Client) Shutdown(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.graph.Shutdown(ctx, in, opts...)
}

type TriggerFunc func(ctx context.Context, trigger *apipb.Interception) (*apipb.Interception, error)

// Trigger is an optional/custom external plugin that when added to a Graphik instance, mutates objects at runtime before & after state changes.
// It should be deployed as a side car to a graphik instance(see Serve())
type Trigger struct {
	fn          TriggerFunc
	expressions []string
}

func NewTrigger(fn TriggerFunc, expressions []string) *Trigger {
	return &Trigger{fn: fn, expressions: expressions}
}

func (t Trigger) Filter(ctx context.Context, _ *empty.Empty) (*apipb.TriggerFilter, error) {
	return &apipb.TriggerFilter{
		Expressions: t.expressions,
	}, nil
}

func (t Trigger) Mutate(ctx context.Context, interception *apipb.Interception) (*apipb.Interception, error) {
	return t.fn(ctx, interception)
}

func (t Trigger) Ping(ctx context.Context, _ *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (t Trigger) Serve(ctx context.Context, cfg *flags.PluginFlags) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	mach := machine.New(ctx)
	defer mach.Close()
	router := http.NewServeMux()
	if cfg.Metrics {
		router.Handle("/metrics", promhttp.Handler())
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &http.Server{
		Handler: router,
	}

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindHTTP)
		if err != nil {
			logger.Error("failed to create http server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failure", zap.Error(err))
		}
	})

	gserver := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger()),
			grpc_validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger.Logger()),
			grpc_validator.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
	)

	apipb.RegisterTriggerServiceServer(gserver, t)
	grpc_prometheus.Register(gserver)

	mach.Go(func(routine machine.Routine) {
		lis, err := net.Listen("tcp", cfg.BindGrpc)
		if err != nil {
			logger.Error("failed to create gRPC server listener", zap.Error(err))
			return
		}
		defer lis.Close()
		if err := gserver.Serve(lis); err != nil && err != http.ErrServerClosed {
			logger.Error("gRPC server failure", zap.Error(err))
		}
	})
	select {
	case <-interrupt:
		mach.Cancel()
		break
	case <-ctx.Done():
		mach.Cancel()
		break
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = server.Shutdown(shutdownCtx)
	stopped := make(chan struct{})
	go func() {
		gserver.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(5 * time.Second)

	select {
	case <-timer.C:
		gserver.Stop()
	case <-stopped:
		timer.Stop()
	}
	mach.Wait()
}
