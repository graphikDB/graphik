package graphik

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/logger"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

type Options struct {
	retry       uint
	tokenSource oauth2.TokenSource
	raftSecret  string
	metrics     bool
	logging     bool
	logPayload  bool
}

type Opt func(o *Options)

func WithRaftSecret(raftSecret string) Opt {
	return func(o *Options) {
		o.raftSecret = raftSecret
	}
}

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

func WithMetrics(metrics bool) Opt {
	return func(o *Options) {
		o.metrics = metrics
	}
}

func WithLogging(logging, logPayload bool) Opt {
	return func(o *Options) {
		o.logging = logging
		o.logPayload = logPayload
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
	if target == "" {
		return nil, errors.New("empty target")
	}
	dialopts := []grpc.DialOption{grpc.WithInsecure()}
	var uinterceptors []grpc.UnaryClientInterceptor
	var sinterceptors []grpc.StreamClientInterceptor
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	uinterceptors = append(uinterceptors, grpc_validator.UnaryClientInterceptor())

	if options.metrics {
		uinterceptors = append(uinterceptors, grpc_prometheus.UnaryClientInterceptor)
		sinterceptors = append(sinterceptors, grpc_prometheus.StreamClientInterceptor)
	}

	if options.tokenSource != nil {
		uinterceptors = append(uinterceptors, unaryAuth(options.tokenSource))
		sinterceptors = append(sinterceptors, streamAuth(options.tokenSource))
	}
	if options.logging {
		lgger := logger.New(true, zap.Bool("client", true))
		uinterceptors = append(uinterceptors, grpc_zap.UnaryClientInterceptor(lgger.Zap()))
		sinterceptors = append(sinterceptors, grpc_zap.StreamClientInterceptor(lgger.Zap()))

		if options.logPayload {
			uinterceptors = append(uinterceptors, grpc_zap.PayloadUnaryClientInterceptor(lgger.Zap(), func(ctx context.Context, fullMethodName string) bool {
				return true
			}))
			sinterceptors = append(sinterceptors, grpc_zap.PayloadStreamClientInterceptor(lgger.Zap(), func(ctx context.Context, fullMethodName string) bool {
				return true
			}))
		}
	}

	uinterceptors = append(uinterceptors, grpc_retry.UnaryClientInterceptor(
		grpc_retry.WithMax(options.retry),
		grpc_retry.WithPerRetryTimeout(1*time.Second),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100*time.Millisecond)),
	))
	dialopts = append(dialopts,
		grpc.WithChainUnaryInterceptor(uinterceptors...),
		grpc.WithChainStreamInterceptor(sinterceptors...),
	)

	conn, err := grpc.DialContext(ctx, target, dialopts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		graph:      apipb.NewDatabaseServiceClient(conn),
		raft:       apipb.NewRaftServiceClient(conn),
		raftSecret: options.raftSecret,
	}, nil
}

type Client struct {
	raftSecret string
	graph      apipb.DatabaseServiceClient
	raft       apipb.RaftServiceClient
}

func toContext(ctx context.Context, tokenSource oauth2.TokenSource) (context.Context, error) {
	token, err := tokenSource.Token()
	if err != nil {
		return ctx, errors.Wrap(err, "failed to get token")
	}
	return metadata.AppendToOutgoingContext(
		ctx,
		"Authorization", fmt.Sprintf("Bearer %v", token.AccessToken),
	), nil
}

// Me returns a Doc of the currently logged in user
func (c *Client) Me(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.Me(ctx, in, opts...)
}

// CreateDoc creates a single doc in the graph
func (c *Client) CreateDoc(ctx context.Context, in *apipb.DocConstructor, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.CreateDoc(ctx, in, opts...)
}

// CreateDocs creates 1-many documents in the graph
func (c *Client) CreateDocs(ctx context.Context, in *apipb.DocConstructors, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.CreateDocs(ctx, in, opts...)
}

// GetDoc gets a doc at the given ref
func (c *Client) GetDoc(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.GetDoc(ctx, in, opts...)
}

// SearchDocs searches for 0-many docs
func (c *Client) SearchDocs(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.SearchDocs(ctx, in, opts...)
}

// EditDoc edites a single doc in the graph
func (c *Client) EditDoc(ctx context.Context, in *apipb.Edit, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.EditDoc(ctx, in, opts...)
}

// EditDocs edites 0-many docs in the graph
func (c *Client) EditDocs(ctx context.Context, in *apipb.EditFilter, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.EditDocs(ctx, in, opts...)
}

// CreateConnection creates a single connection in the graph
func (c *Client) CreateConnection(ctx context.Context, in *apipb.ConnectionConstructor, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.CreateConnection(ctx, in, opts...)
}

// CreateConnections creates 1-many connections in the graph
func (c *Client) CreateConnections(ctx context.Context, in *apipb.ConnectionConstructors, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.CreateConnections(ctx, in, opts...)
}

// GetConnection gets a connection at the given ref
func (c *Client) GetConnection(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.GetConnection(ctx, in, opts...)
}

// SearchConnections searches for 0-many connections
func (c *Client) SearchConnections(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.SearchConnections(ctx, in, opts...)
}

// EditConnection edites a single connection in the graph
func (c *Client) EditConnection(ctx context.Context, in *apipb.Edit, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.EditConnection(ctx, in, opts...)
}

// EditConnections edites 0-many connections in the graph
func (c *Client) EditConnections(ctx context.Context, in *apipb.EditFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.EditConnections(ctx, in, opts...)
}

// ConnectionsFrom returns connections from the given doc that pass the filter
func (c *Client) ConnectionsFrom(ctx context.Context, in *apipb.ConnectFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsFrom(ctx, in, opts...)
}

// ConnectionsTo returns connections to the given doc that pass the filter
func (c *Client) ConnectionsTo(ctx context.Context, in *apipb.ConnectFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.ConnectionsTo(ctx, in, opts...)
}

// Broadcast broadcasts a message to a pubsub channel
func (c *Client) Broadcast(ctx context.Context, in *apipb.OutboundMessage, opts ...grpc.CallOption) error {
	_, err := c.graph.Broadcast(ctx, in, opts...)
	return err
}

// Stream opens a stream of messages that pass a filter on a pubsub channel
func (c *Client) Stream(ctx context.Context, in *apipb.StreamFilter, handler func(msg *apipb.Message) bool, opts ...grpc.CallOption) error {
	stream, err := c.graph.Stream(ctx, in, opts...)
	if err != nil {
		return errors.Wrapf(err, "failed to create client stream")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := stream.Recv()
			if err != nil {
				return errors.Wrap(err, "failed to receive message")
			}
			if !handler(msg) {
				return nil
			}
		}
	}
}

// PushDocConstructors streams doc constructors to & from the graph as they are created
func (c *Client) PushDocConstructors(ctx context.Context, ch <-chan *apipb.DocConstructor, opts ...grpc.CallOption) error {
	stream, err := c.graph.PushDocConstructors(ctx, opts...)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

// PushConnectionConstructors streams connection constructors to & from the graph as they are created
func (c *Client) PushConnectionConstructors(ctx context.Context, ch <-chan *apipb.ConnectionConstructor, opts ...grpc.CallOption) error {
	stream, err := c.graph.PushConnectionConstructors(ctx, opts...)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

// GetSchema gets information about node/connection types, type-validators, indexes, and authorizers
func (c *Client) GetSchema(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Schema, error) {
	return c.graph.GetSchema(ctx, in, opts...)
}

// SetIndexes sets all of the indexes in the graph
func (c *Client) SetIndexes(ctx context.Context, in *apipb.Indexes, opts ...grpc.CallOption) error {
	_, err := c.graph.SetIndexes(ctx, in, opts...)
	return err
}

// SetAuthorizers sets all of the authorizers in the graph
func (c *Client) SetAuthorizers(ctx context.Context, in *apipb.Authorizers, opts ...grpc.CallOption) error {
	_, err := c.graph.SetAuthorizers(ctx, in, opts...)
	return err
}

// SetConstraints sets all of the type validators in the graph
func (c *Client) SetConstraints(ctx context.Context, in *apipb.Constraints, opts ...grpc.CallOption) error {
	_, err := c.graph.SetConstraints(ctx, in, opts...)
	return err
}

// SetTriggers sets all of the triggers in the graph
func (c *Client) SetTriggers(ctx context.Context, in *apipb.Triggers, opts ...grpc.CallOption) error {
	_, err := c.graph.SetTriggers(ctx, in, opts...)
	return err
}

// SeedDocs
func (c *Client) SeedDocs(ctx context.Context, docChan <-chan *apipb.Doc, opts ...grpc.CallOption) error {
	stream, err := c.graph.SeedDocs(ctx, opts...)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-docChan:
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

// SeedConnections
func (c *Client) SeedConnections(ctx context.Context, connectionChan <-chan *apipb.Connection, opts ...grpc.CallOption) error {
	stream, err := c.graph.SeedConnections(ctx, opts...)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-connectionChan:
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

// SearchAndConnect searches for documents and forms connections based on whether they pass a filter
func (c *Client) SearchAndConnect(ctx context.Context, in *apipb.SearchConnectFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.SearchAndConnect(ctx, in, opts...)
}

// SearchAndConnectMe searches for documents and forms connections between the origin user & the document based on whether they pass a filter
func (c *Client) SearchAndConnectMe(ctx context.Context, in *apipb.SearchConnectMeFilter, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.SearchAndConnectMe(ctx, in, opts...)
}

// Traverse searches for 0-many docs using a graph traversal algorithm
func (c *Client) Traverse(ctx context.Context, in *apipb.TraverseFilter, opts ...grpc.CallOption) (*apipb.Traversals, error) {
	return c.graph.Traverse(ctx, in, opts...)
}

// DelDoc deletes a doc by reference
func (c *Client) DelDoc(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) error {
	_, err := c.graph.DelDoc(ctx, in, opts...)
	return err
}

// DelDocs deletes 0-many docs that pass a Filter
func (c *Client) DelDocs(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) error {
	_, err := c.graph.DelDocs(ctx, in, opts...)
	return err
}

// ExistsDoc checks if a doc exists in the graph
func (c *Client) ExistsDoc(ctx context.Context, in *apipb.ExistsFilter, opts ...grpc.CallOption) (*apipb.Boolean, error) {
	return c.graph.ExistsDoc(ctx, in, opts...)
}

// ExistsConnection checks if a connection exists in the graph
func (c *Client) ExistsConnection(ctx context.Context, in *apipb.ExistsFilter, opts ...grpc.CallOption) (*apipb.Boolean, error) {
	return c.graph.ExistsConnection(ctx, in, opts...)
}

// HasDoc checks if a doc exists in the graph by reference
func (c *Client) HasDoc(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) (*apipb.Boolean, error) {
	return c.graph.HasDoc(ctx, in, opts...)
}

// HasConnection checks if a connection exists in the graph by reference
func (c *Client) HasConnection(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) (*apipb.Boolean, error) {
	return c.graph.HasConnection(ctx, in, opts...)
}

// DelConnection deletes a single connection that pass a Filter
func (c *Client) DelConnection(ctx context.Context, in *apipb.Ref, opts ...grpc.CallOption) error {
	_, err := c.graph.DelConnection(ctx, in, opts...)
	return err
}

// DelConnections deletes 0-many connections that pass a Filter
func (c *Client) DelConnections(ctx context.Context, in *apipb.Filter, opts ...grpc.CallOption) error {
	_, err := c.graph.DelConnections(ctx, in, opts...)
	return err
}

// AggregateDocs executes an aggregation function against a set of documents
func (c *Client) AggregateDocs(ctx context.Context, in *apipb.AggFilter, opts ...grpc.CallOption) (*apipb.Number, error) {
	return c.graph.AggregateDocs(ctx, in, opts...)
}

// AggregateConnections executes an aggregation function against a set of connections
func (c *Client) AggregateConnections(ctx context.Context, in *apipb.AggFilter, opts ...grpc.CallOption) (*apipb.Number, error) {
	return c.graph.AggregateConnections(ctx, in, opts...)
}

// AddPeer adds a peer node to the raft cluster.
func (c *Client) JoinCluster(ctx context.Context, peer *apipb.Peer, opts ...grpc.CallOption) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-graphik-raft-secret", c.raftSecret)
	_, err := c.raft.JoinCluster(ctx, peer, opts...)
	return err
}

// ClusterState returns information about the raft cluster
func (c *Client) ClusterState(ctx context.Context, _ *empty.Empty, opts ...grpc.CallOption) error {
	_, err := c.raft.ClusterState(ctx, &empty.Empty{}, opts...)
	return err
}

// Ping checks if the server is healthy.
func (c *Client) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*apipb.Pong, error) {
	return c.raft.Ping(ctx, in, opts...)
}

// PutDocs puts a batch of docs in the graph
func (c *Client) PutDoc(ctx context.Context, in *apipb.Doc, opts ...grpc.CallOption) (*apipb.Doc, error) {
	return c.graph.PutDoc(ctx, in, opts...)
}

// PutDoc creates a Doc if it doesnt exist already, otherwise it replaces it
func (c *Client) PutDocs(ctx context.Context, in *apipb.Docs, opts ...grpc.CallOption) (*apipb.Docs, error) {
	return c.graph.PutDocs(ctx, in, opts...)
}

// PutConnection create-or-replaces a Connection in the graph
func (c *Client) PutConnection(ctx context.Context, in *apipb.Connection, opts ...grpc.CallOption) (*apipb.Connection, error) {
	return c.graph.PutConnection(ctx, in, opts...)
}

// PutConnections puts a batch of connections in the graph
func (c *Client) PutConnections(ctx context.Context, in *apipb.Connections, opts ...grpc.CallOption) (*apipb.Connections, error) {
	return c.graph.PutConnections(ctx, in, opts...)
}
