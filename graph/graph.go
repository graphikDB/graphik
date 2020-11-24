package graph

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/auth"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/sortable"
	"github.com/autom8ter/graphik/vm"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/google/uuid"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode    = 0600
	changeChannel = "changes"
)

var (
	DONE = errors.New("DONE")
	// Bucket names we perform transactions in
	dbEdges = []byte("edges")
	dbNodes = []byte("nodes")
	// An error indicating a given key does not exist
	ErrNotFound = errors.New("not found")
)

type GraphStore struct {
	// db is the underlying handle to the db.
	db *bbolt.DB
	// The path to the Bolt database file
	path        string
	mu          sync.RWMutex
	triggerMu   sync.RWMutex
	edgesTo     map[string][]*apipb.Path
	edgesFrom   map[string][]*apipb.Path
	machine     *machine.Machine
	triggers    []*triggerClient
	authorizers []cel.Program
	auth        *auth.Config
	closers     []func()
	closeOnce   sync.Once
	initOnce    sync.Once
}

// NewGraphStore takes a file path and returns a connected Raft backend.
func NewGraphStore(ctx context.Context, flgs *flags.Flags) (*GraphStore, error) {
	os.MkdirAll(flgs.StoragePath, 0700)
	path := filepath.Join(flgs.StoragePath, "graph.db")
	handle, err := bbolt.Open(path, dbFileMode, nil)
	if err != nil {
		return nil, err
	}
	var triggers []*triggerClient
	var closers []func()
	for _, plugin := range flgs.Triggers {
		if plugin == "" {
			continue
		}
		tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		conn, err := grpc.DialContext(
			tctx,
			plugin,
			grpc.WithChainUnaryInterceptor(grpc_zap.UnaryClientInterceptor(logger.Logger())),
			grpc.WithInsecure(),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial trigger")
		}
		trig := apipb.NewTriggerServiceClient(conn)
		matchers, err := trig.Filter(ctx, &empty.Empty{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial trigger")
		}
		programs, err := vm.Programs(matchers.GetExpressions())
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile trigger matchers")
		}
		triggers = append(triggers, &triggerClient{
			TriggerServiceClient: trig,
			matchers:             programs,
		})
		closers = append(closers, func() {
			conn.Close()
		})
	}
	var programs []cel.Program
	if len(flgs.Authorizers) > 0 && flgs.Authorizers[0] != "" {
		programs, err = vm.Programs(flgs.Authorizers)
		if err != nil {
			return nil, err
		}
	}
	jwks, err := auth.New(flgs.JWKS)
	if err != nil {
		return nil, err
	}
	g := &GraphStore{
		db:          handle,
		path:        path,
		mu:          sync.RWMutex{},
		edgesTo:     map[string][]*apipb.Path{},
		edgesFrom:   map[string][]*apipb.Path{},
		machine:     machine.New(ctx),
		triggers:    triggers,
		authorizers: programs,
		auth:        jwks,
		closers:     closers,
		closeOnce:   sync.Once{},
		initOnce:    sync.Once{},
	}
	if err := g.init(ctx); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GraphStore) Ping(ctx context.Context, e *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (g *GraphStore) GetSchema(ctx context.Context, _ *empty.Empty) (*apipb.Schema, error) {
	e, err := g.EdgeTypes(ctx)
	if err != nil {
		return nil, err
	}
	n, err := g.NodeTypes(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Schema{
		EdgeTypes: e,
		NodeTypes: n,
	}, nil
}

func (g *GraphStore) Me(ctx context.Context, filter *apipb.MeFilter) (*apipb.NodeDetail, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	detail := &apipb.NodeDetail{
		Path:       identity.Path,
		Attributes: identity.Attributes,
		Metadata:   identity.Metadata,
		EdgesTo:    &apipb.EdgeDetails{},
		EdgesFrom:  &apipb.EdgeDetails{},
	}
	if filter.EdgesFrom != nil {
		from, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
			NodePath:    identity.GetPath(),
			Gtype:       filter.GetEdgesFrom().GetGtype(),
			Expressions: filter.GetEdgesFrom().GetExpressions(),
			Limit:       filter.GetEdgesFrom().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, f := range from.GetEdges() {
			toNode, err := g.GetNode(ctx, f.To)
			if err != nil {
				return nil, err
			}
			fromNode, err := g.GetNode(ctx, f.From)
			if err != nil {
				return nil, err
			}
			edetail := &apipb.EdgeDetail{
				Path:       f.Path,
				Attributes: f.Attributes,

				From:     fromNode,
				To:       toNode,
				Metadata: f.Metadata,
			}
			detail.EdgesFrom.Edges = append(detail.EdgesFrom.Edges, edetail)
		}
	}
	if filter.EdgesTo != nil {
		from, err := g.EdgesTo(ctx, &apipb.EdgeFilter{
			NodePath:    identity.GetPath(),
			Gtype:       filter.GetEdgesFrom().GetGtype(),
			Expressions: filter.GetEdgesFrom().GetExpressions(),
			Limit:       filter.GetEdgesFrom().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, f := range from.GetEdges() {
			toNode, err := g.GetNode(ctx, f.To)
			if err != nil {
				return nil, err
			}
			fromNode, err := g.GetNode(ctx, f.From)
			if err != nil {
				return nil, err
			}
			edetail := &apipb.EdgeDetail{
				Path:       f.Path,
				Attributes: f.Attributes,

				From:     fromNode,
				To:       toNode,
				Metadata: f.Metadata,
			}
			detail.EdgesTo.Edges = append(detail.EdgesTo.Edges, edetail)
		}
	}
	return detail, nil
}

func (g *GraphStore) CreateNodes(ctx context.Context, constructors *apipb.NodeConstructors) (*apipb.Nodes, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := timestamppb.Now()
	method := g.MethodContext(ctx)
	var nodes []*apipb.Node
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		nodeBucket := tx.Bucket(dbNodes)
		for _, constructor := range constructors.GetNodes() {
			bucket := nodeBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
			if constructor.Path.GetGid() == "" {
				constructor.Path.Gid = uuid.New().String()
			}
			node := &apipb.Node{
				Path:       constructor.GetPath(),
				Attributes: constructor.GetAttributes(),
				Metadata: &apipb.Metadata{
					CreatedAt: now,
					UpdatedAt: now,
					UpdatedBy: identity.GetPath(),
				},
			}
			if bucket == nil {
				bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			seq, _ := bucket.NextSequence()
			node.Metadata.Sequence = seq
			nodes = append(nodes, node)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	nodess, err := g.SetNodes(ctx, nodes...)
	if err != nil {
		return nil, err
	}
	var changes []*apipb.NodeChange
	for _, node := range nodes {
		changes = append(changes, &apipb.NodeChange{
			After: node,
		})
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      method,
		Identity:    identity,
		Timestamp:   now,
		NodeChanges: changes,
	})
	nodess.Sort()
	return nodess, nil
}

func (g *GraphStore) CreateEdge(ctx context.Context, constructor *apipb.EdgeConstructor) (*apipb.Edge, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	edges, err := g.CreateEdges(ctx, &apipb.EdgeConstructors{Edges: []*apipb.EdgeConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return edges.GetEdges()[0], nil
}

func (g *GraphStore) CreateEdges(ctx context.Context, constructors *apipb.EdgeConstructors) (*apipb.Edges, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := timestamppb.Now()
	method := g.MethodContext(ctx)
	var edges = []*apipb.Edge{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		edgeBucket := tx.Bucket(dbEdges)
		for _, constructor := range constructors.GetEdges() {
			bucket := edgeBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = edgeBucket.CreateBucketIfNotExists([]byte(constructor.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if constructor.Path.GetGid() == "" {
				constructor.Path.Gid = uuid.New().String()
			}
			edge := &apipb.Edge{
				Path:       constructor.GetPath(),
				Attributes: constructor.GetAttributes(),
				Metadata: &apipb.Metadata{
					CreatedAt: now,
					UpdatedAt: now,
					UpdatedBy: identity.GetPath(),
				},
				Directed: constructor.Directed,
				From:     constructor.GetFrom(),
				To:       constructor.GetTo(),
			}
			seq, _ := bucket.NextSequence()
			edge.Metadata.Sequence = seq
			edges = append(edges, edge)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	edgess, err := g.SetEdges(ctx, edges...)
	if err != nil {
		return nil, err
	}
	var changes []*apipb.EdgeChange
	for _, node := range edges {
		changes = append(changes, &apipb.EdgeChange{
			After: node,
		})
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      method,
		Identity:    identity,
		Timestamp:   now,
		EdgeChanges: changes,
	})
	edgess.Sort()
	return edgess, nil
}

func (g *GraphStore) Publish(ctx context.Context, message *apipb.OutboundMessage) (*empty.Empty, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	return &empty.Empty{}, g.machine.PubSub().Publish(message.Channel, &apipb.Message{
		Channel:   message.Channel,
		Data:      message.Data,
		Sender:    identity.GetPath(),
		Timestamp: timestamppb.Now(),
	})
}

func (g *GraphStore) Subscribe(filter *apipb.ChannelFilter, server apipb.GraphService_SubscribeServer) error {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		val, ok := msg.(*apipb.Message)
		if !ok {
			logger.Error("invalid message type received during subscription")
			return false
		}
		result, err := vm.Eval(programs, val)
		if err != nil {
			logger.Error("subscription filter failure", zap.Error(err))
			return false
		}
		return result
	}
	if err := g.machine.PubSub().SubscribeFilter(server.Context(), filter.Channel, filterFunc, func(msg interface{}) {
		if err, ok := msg.(error); ok && err != nil {
			logger.Error("failed to send subscription", zap.Error(err))
			return
		}
		if val, ok := msg.(*apipb.Message); ok && val != nil {
			if err := server.Send(val); err != nil {
				logger.Error("failed to send subscription", zap.Error(err))
				return
			}
		}
	}); err != nil {
		return err
	}
	return nil
}

func (g *GraphStore) SubscribeChanges(filter *apipb.ExpressionFilter, server apipb.GraphService_SubscribeChangesServer) error {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		val, ok := msg.(*apipb.Change)
		if !ok {
			logger.Error("invalid message type received during change subscription")
			return false
		}
		result, err := vm.Eval(programs, val)
		if err != nil {
			logger.Error("subscription change failure", zap.Error(err))
			return false
		}
		return result
	}
	if err := g.machine.PubSub().SubscribeFilter(server.Context(), changeChannel, filterFunc, func(msg interface{}) {
		if err, ok := msg.(error); ok && err != nil {
			logger.Error("failed to send change", zap.Error(err))
			return
		}
		if val, ok := msg.(*apipb.Change); ok && val != nil {
			if err := server.Send(val); err != nil {
				logger.Error("failed to send change", zap.Error(err))
				return
			}
		}
	}); err != nil {
		return err
	}
	return nil
}

func (r *GraphStore) Export(ctx context.Context, _ *empty.Empty) (*apipb.Graph, error) {
	identity := r.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	nodes, err := r.AllNodes(ctx)
	if err != nil {
		return nil, err
	}
	edges, err := r.AllEdges(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *GraphStore) Import(ctx context.Context, graph *apipb.Graph) (*apipb.Graph, error) {
	identity := r.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	nodes, err := r.SetNodes(ctx, graph.GetNodes().GetNodes()...)
	if err != nil {
		return nil, err
	}
	edges, err := r.SetEdges(ctx, graph.GetEdges().GetEdges()...)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (g *GraphStore) Shutdown(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	go g.Close()
	return &empty.Empty{}, nil
}

// Close is used to gracefully close the Database.
func (b *GraphStore) Close() {
	b.closeOnce.Do(func() {
		b.machine.Close()
		for _, closer := range b.closers {
			closer()
		}
		b.machine.Wait()
		if err := b.db.Close(); err != nil {
			logger.Error("failed to close db", zap.Error(err))
		}
	})
}

func (g *GraphStore) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var edge apipb.Edge
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		bits := bucket.Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, &edge); err != nil {
			return err
		}
		return nil
	}); err != nil && err != DONE {
		return nil, err
	}
	return &edge, nil
}

func (g *GraphStore) RangeSeekEdges(ctx context.Context, gType, seek string, fn func(e *apipb.Edge) bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if gType == apipb.Any {
		types, err := g.EdgeTypes(ctx)
		if err != nil {
			return err
		}
		for _, edgeType := range types {
			if edgeType == apipb.Any {
				continue
			}
			if err := g.RangeSeekEdges(ctx, edgeType, seek, fn); err != nil {
				return err
			}
		}
		return nil
	}
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges).Bucket([]byte(gType))
		if bucket == nil {
			return ErrNotFound
		}
		c := bucket.Cursor()
		for k, v := c.Seek([]byte(seek)); k != nil; k, v = c.Next() {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var edge apipb.Edge
			if err := proto.Unmarshal(v, &edge); err != nil {
				return err
			}
			if !fn(&edge) {
				return DONE
			}
		}
		return nil
	}); err != nil && err != DONE {
		return err
	}
	return nil
}

func (n *GraphStore) AllNodes(ctx context.Context) (*apipb.Nodes, error) {
	identity := n.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var nodes []*apipb.Node
	if err := n.RangeNodes(ctx, apipb.Any, func(node *apipb.Node) bool {
		nodes = append(nodes, node)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Nodes{
		Nodes: nodes,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (g *GraphStore) RangeEdges(ctx context.Context, gType string, fn func(n *apipb.Edge) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if gType == apipb.Any {
			types, err := g.EdgeTypes(ctx)
			if err != nil {
				return err
			}
			for _, edgeType := range types {
				if err := g.RangeEdges(ctx, edgeType, fn); err != nil {
					return err
				}
			}
			return nil
		}
		bucket := tx.Bucket(dbEdges).Bucket([]byte(gType))
		if bucket == nil {
			return ErrNotFound
		}

		return bucket.ForEach(func(k, v []byte) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var edge apipb.Edge
			if err := proto.Unmarshal(v, &edge); err != nil {
				return err
			}
			if !fn(&edge) {
				return DONE
			}
			return nil
		})
	}); err != nil && err != DONE {
		return err
	}
	return nil
}

func (g *GraphStore) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var node apipb.Node
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		bits := bucket.Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, &node); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &node, nil
}

func (g *GraphStore) RangeNodes(ctx context.Context, gType string, fn func(n *apipb.Node) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.db.View(func(tx *bbolt.Tx) error {
		if gType == apipb.Any {
			types, err := g.NodeTypes(ctx)
			if err != nil {
				return err
			}
			for _, nodeType := range types {
				if err := g.RangeNodes(ctx, nodeType, fn); err != nil {
					return err
				}
			}
			return nil
		}
		bucket := tx.Bucket(dbNodes).Bucket([]byte(gType))
		if bucket == nil {
			return ErrNotFound
		}

		return bucket.ForEach(func(k, v []byte) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var node apipb.Node
			if err := proto.Unmarshal(v, &node); err != nil {
				return err
			}
			if !fn(&node) {
				return DONE
			}
			return nil
		})
	}); err != nil && err != DONE {
		return err
	}
	return nil
}

func (g *GraphStore) createIdentity(ctx context.Context, constructor *apipb.NodeConstructor) (*apipb.Node, error) {
	now := timestamppb.Now()
	var node *apipb.Node
	var err error
	if constructor.Path.Gid == "" {
		constructor.Path.Gid = uuid.New().String()
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		node = &apipb.Node{
			Path:       constructor.GetPath(),
			Attributes: constructor.GetAttributes(),
			Metadata: &apipb.Metadata{
				CreatedAt: now,
				UpdatedAt: now,
				UpdatedBy: constructor.GetPath(),
			},
		}
		nodeBucket := tx.Bucket(dbNodes)
		bucket := nodeBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
		if constructor.Path.GetGid() == "" {
			constructor.Path.Gid = uuid.New().String()
		}
		if bucket == nil {
			bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
			if err != nil {
				return err
			}
		}
		seq, _ := bucket.NextSequence()
		node.Metadata.Sequence = seq
		node, err = g.SetNode(ctx, node)
		return err
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (g *GraphStore) CreateNode(ctx context.Context, constructor *apipb.NodeConstructor) (*apipb.Node, error) {
	nodes, err := g.CreateNodes(ctx, &apipb.NodeConstructors{Nodes: []*apipb.NodeConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return nodes.GetNodes()[0], nil
}

func (g *GraphStore) SetNode(ctx context.Context, node *apipb.Node) (*apipb.Node, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	nodes, err := g.SetNodes(ctx, node)
	if err != nil {
		return nil, err
	}
	return nodes.GetNodes()[0], nil
}

func (g *GraphStore) SetNodes(ctx context.Context, nodes ...*apipb.Node) (*apipb.Nodes, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, node := range nodes {
			if err := ctx.Err(); err != nil {
				return err
			}
			if node.GetMetadata().GetCreatedAt() == nil {
				node.GetMetadata().CreatedAt = timestamppb.Now()
			}
			if node.GetMetadata().GetUpdatedAt() == nil {
				node.GetMetadata().UpdatedAt = timestamppb.Now()
			}
			node.GetMetadata().Version += 1
			node.GetMetadata().Hash = fmt.Sprintf("%x", node.String())
			bits, err := proto.Marshal(node)
			if err != nil {
				return err
			}
			nodeBucket := tx.Bucket(dbNodes)
			bucket := nodeBucket.Bucket([]byte(node.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if err := bucket.Put([]byte(node.GetPath().GetGid()), bits); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	nds := &apipb.Nodes{Nodes: nodes}
	nds.Sort()
	return nds, nil
}

func (g *GraphStore) SetEdge(ctx context.Context, edge *apipb.Edge) (*apipb.Edge, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	edges, err := g.SetEdges(ctx, edge)
	if err != nil {
		return nil, err
	}
	return edges.GetEdges()[0], nil
}

func (g *GraphStore) SetEdges(ctx context.Context, edges ...*apipb.Edge) (*apipb.Edges, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, edge := range edges {
			nodeBucket := tx.Bucket(dbNodes)
			{
				fromBucket := nodeBucket.Bucket([]byte(edge.From.Gtype))
				if fromBucket == nil {
					return errors.Errorf("from node %s does not exist", edge.From.String())
				}
				val := fromBucket.Get([]byte(edge.From.Gid))
				if val == nil {
					return errors.Errorf("from node %s does not exist", edge.From.String())
				}
			}
			{
				toBucket := nodeBucket.Bucket([]byte(edge.To.Gtype))
				if toBucket == nil {
					return errors.Errorf("to node %s does not exist", edge.To.String())
				}
				val := toBucket.Get([]byte(edge.To.Gid))
				if val == nil {
					return errors.Errorf("to node %s does not exist", edge.To.String())
				}
			}
			if edge.GetMetadata().GetCreatedAt() == nil {
				edge.GetMetadata().CreatedAt = timestamppb.Now()
			}
			if edge.GetMetadata().GetUpdatedAt() == nil {
				edge.GetMetadata().UpdatedAt = timestamppb.Now()
			}
			edge.GetMetadata().Version += 1
			edge.GetMetadata().Hash = fmt.Sprintf("%x", edge.String())
			bits, err := proto.Marshal(edge)
			if err != nil {
				return err
			}
			edgeBucket := tx.Bucket(dbEdges)
			edgeBucket = edgeBucket.Bucket([]byte(edge.GetPath().GetGtype()))
			if edgeBucket == nil {
				edgeBucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if err := edgeBucket.Put([]byte(edge.GetPath().GetGid()), bits); err != nil {
				return err
			}
			g.mu.Lock()
			defer g.mu.Unlock()
			g.edgesFrom[edge.GetFrom().String()] = append(g.edgesFrom[edge.GetFrom().String()], edge.GetPath())
			g.edgesTo[edge.GetTo().String()] = append(g.edgesTo[edge.GetTo().String()], edge.GetPath())
			if !edge.Directed {
				g.edgesTo[edge.GetFrom().String()] = append(g.edgesTo[edge.GetFrom().String()], edge.GetPath())
				g.edgesFrom[edge.GetTo().String()] = append(g.edgesFrom[edge.GetTo().String()], edge.GetPath())
			}
			sortPaths(g.edgesFrom[edge.GetFrom().String()])
			sortPaths(g.edgesTo[edge.GetTo().String()])
		}
		return nil
	}); err != nil {
		return nil, err
	}
	e := &apipb.Edges{Edges: edges}
	e.Sort()
	return e, nil
}

func (n *GraphStore) PatchNode(ctx context.Context, value *apipb.Patch) (*apipb.Node, error) {
	identity := n.NodeContext(ctx)
	node, err := n.GetNode(ctx, value.GetPath())
	if err != nil {
		return nil, err
	}
	change := &apipb.NodeChange{
		Before: node,
	}
	for k, v := range value.GetAttributes().GetFields() {
		node.Attributes.GetFields()[k] = v
	}
	node.GetMetadata().UpdatedAt = timestamppb.Now()
	node.GetMetadata().UpdatedBy = identity.GetPath()
	change.After = node
	node, err = n.SetNode(ctx, node)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      n.MethodContext(ctx),
		Identity:    identity,
		Timestamp:   node.Metadata.UpdatedAt,
		EdgeChanges: nil,
		NodeChanges: []*apipb.NodeChange{change},
	})
	return node, nil
}

func (n *GraphStore) PatchNodes(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Nodes, error) {
	identity := n.NodeContext(ctx)
	var changes []*apipb.NodeChange
	var nodes []*apipb.Node
	method := n.MethodContext(ctx)
	now := timestamppb.Now()
	before, err := n.SearchNodes(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, node := range before.GetNodes() {
		change := &apipb.NodeChange{
			Before: node,
		}
		for k, v := range patch.GetPatch().GetAttributes().GetFields() {
			node.Attributes.GetFields()[k] = v
		}
		node.GetMetadata().UpdatedAt = now
		node.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = node
		nodes = append(nodes, node)
		changes = append(changes, change)
	}

	nodess, err := n.SetNodes(ctx, nodes...)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      method,
		Identity:    identity,
		Timestamp:   now,
		EdgeChanges: nil,
		NodeChanges: changes,
	})
	return nodess, nil
}

func (g *GraphStore) EdgeTypes(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var types []string
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbEdges).ForEach(func(name []byte, _ []byte) error {
			types = append(types, string(name))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	sort.Strings(types)
	return types, nil
}

func (g *GraphStore) NodeTypes(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var types []string
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbNodes).ForEach(func(name []byte, _ []byte) error {
			types = append(types, string(name))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	sort.Strings(types)
	return types, nil
}

func (g *GraphStore) RangeFrom(ctx context.Context, path *apipb.Path, fn func(e *apipb.Edge) bool) error {
	g.mu.RLock()
	paths := g.edgesFrom[path.String()]
	g.mu.RUnlock()
	for _, path := range paths {
		edge, err := g.GetEdge(ctx, path)
		if err != nil {
			return err
		}
		if !fn(edge) {
			return nil
		}
	}
	return nil
}

func (g *GraphStore) RangeTo(ctx context.Context, path *apipb.Path, fn func(edge *apipb.Edge) bool) error {
	g.mu.RLock()
	paths := g.edgesTo[path.String()]
	g.mu.RUnlock()
	for _, edge := range paths {
		e, err := g.GetEdge(ctx, edge)
		if err != nil {
			return err
		}
		if !fn(e) {
			return nil
		}
	}
	return nil
}

func (g *GraphStore) EdgesFrom(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err = g.RangeFrom(ctx, filter.NodePath, func(edge *apipb.Edge) bool {
		if filter.Gtype != "*" {
			if edge.GetPath().GetGtype() != filter.Gtype {
				return true
			}
		}

		pass, err = vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, err
}

func (n *GraphStore) HasNode(ctx context.Context, path *apipb.Path) bool {
	node, _ := n.GetNode(ctx, path)
	return node != nil
}

func (n *GraphStore) HasEdge(ctx context.Context, path *apipb.Path) bool {
	edge, _ := n.GetEdge(ctx, path)
	return edge != nil
}

func (n *GraphStore) SearchNodes(ctx context.Context, filter *apipb.Filter) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	if err := n.RangeNodes(ctx, filter.Gtype, func(node *apipb.Node) bool {
		pass, err := vm.Eval(programs, node)
		if err != nil {
			return true
		}
		if pass {
			nodes = append(nodes, node)
		}
		return len(nodes) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Nodes{
		Nodes: nodes,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *GraphStore) FilterNode(ctx context.Context, nodeType string, filter func(node *apipb.Node) bool) (*apipb.Nodes, error) {
	var filtered []*apipb.Node
	if err := n.RangeNodes(ctx, nodeType, func(node *apipb.Node) bool {
		if filter(node) {
			filtered = append(filtered, node)
		}
		return true
	}); err != nil {
		return nil, err
	}
	toreturn := &apipb.Nodes{
		Nodes: filtered,
	}
	toreturn.Sort()
	return toreturn, nil
}

func (e *GraphStore) EdgesTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err := e.RangeTo(ctx, filter.NodePath, func(edge *apipb.Edge) bool {
		if filter.Gtype != apipb.Any {
			if edge.GetPath().GetGtype() != filter.Gtype {
				return true
			}
		}
		pass, err = vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *GraphStore) AllEdges(ctx context.Context) (*apipb.Edges, error) {
	var edges []*apipb.Edge
	if err := n.RangeEdges(ctx, apipb.Any, func(edge *apipb.Edge) bool {
		edges = append(edges, edge)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (n *GraphStore) PatchEdge(ctx context.Context, value *apipb.Patch) (*apipb.Edge, error) {
	identity := n.NodeContext(ctx)
	edge, err := n.GetEdge(ctx, value.GetPath())
	if err != nil {
		return nil, err
	}
	change := &apipb.EdgeChange{
		Before: edge,
	}
	for k, v := range value.GetAttributes().GetFields() {
		edge.Attributes.GetFields()[k] = v
	}
	edge.GetMetadata().UpdatedAt = timestamppb.Now()
	edge.GetMetadata().UpdatedBy = identity.GetPath()
	change.After = edge
	edge, err = n.SetEdge(ctx, edge)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      n.MethodContext(ctx),
		Identity:    identity,
		Timestamp:   edge.Metadata.UpdatedAt,
		EdgeChanges: []*apipb.EdgeChange{change},
	})
	return edge, nil
}

func (n *GraphStore) PatchEdges(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Edges, error) {
	identity := n.NodeContext(ctx)
	var changes []*apipb.EdgeChange
	var edges []*apipb.Edge
	method := n.MethodContext(ctx)
	now := timestamppb.Now()
	before, err := n.SearchEdges(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, edge := range before.GetEdges() {
		change := &apipb.EdgeChange{
			Before: edge,
		}
		for k, v := range patch.GetPatch().GetAttributes().GetFields() {
			edge.Attributes.GetFields()[k] = v
		}
		edge.GetMetadata().UpdatedAt = now
		edge.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = edge
		edges = append(edges, edge)
		changes = append(changes, change)
	}

	edgess, err := n.SetEdges(ctx, edges...)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:      method,
		Identity:    identity,
		Timestamp:   now,
		EdgeChanges: changes,
	})
	return edgess, nil
}

func (e *GraphStore) SearchEdges(ctx context.Context, filter *apipb.Filter) (*apipb.Edges, error) {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	if err := e.RangeEdges(ctx, filter.Gtype, func(edge *apipb.Edge) bool {
		pass, err := vm.Eval(programs, edge)
		if err != nil {
			return true
		}
		if pass {
			edges = append(edges, edge)
		}
		return len(edges) < int(filter.Limit)
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, nil
}

func (g *GraphStore) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	graph := &apipb.Graph{
		Nodes: &apipb.Nodes{},
		Edges: &apipb.Edges{},
	}
	nodes, err := g.SearchNodes(ctx, filter.Nodes)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.GetNodes() {
		graph.Nodes.Nodes = append(graph.Nodes.Nodes, node)
		edges, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdges().GetGtype(),
			Expressions: filter.GetEdges().GetExpressions(),
			Limit:       filter.GetEdges().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		graph.Edges.Edges = append(graph.Edges.Edges, edges.GetEdges()...)
	}
	graph.Edges.Sort()
	graph.Nodes.Sort()
	return graph, err
}

func (g *GraphStore) GetEdgeDetail(ctx context.Context, path *apipb.Path) (*apipb.EdgeDetail, error) {
	e, err := g.GetEdge(ctx, path)
	if err != nil {
		return nil, err
	}
	from, err := g.GetNode(ctx, e.From)
	if err != nil {
		return nil, err
	}
	to, err := g.GetNode(ctx, e.To)
	if err != nil {
		return nil, err
	}
	return &apipb.EdgeDetail{
		Path:       e.GetPath(),
		Attributes: e.GetAttributes(),
		Directed:   e.GetDirected(),
		From:       from,
		To:         to,
		Metadata:   e.GetMetadata(),
	}, nil
}

func (g *GraphStore) GetNodeDetail(ctx context.Context, filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	detail := &apipb.NodeDetail{
		Path:      filter.GetPath(),
		EdgesTo:   &apipb.EdgeDetails{},
		EdgesFrom: &apipb.EdgeDetails{},
	}
	node, err := g.GetNode(ctx, filter.GetPath())
	if err != nil {
		return nil, err
	}
	detail.Metadata = node.Metadata
	detail.Attributes = node.Attributes
	if val := filter.GetFromEdges(); val != nil {
		edges, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       val.GetGtype(),
			Expressions: val.GetExpressions(),
			Limit:       val.GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edges.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesFrom.Edges = append(detail.EdgesFrom.Edges, eDetail)
		}
	}

	if val := filter.GetToEdges(); val != nil {
		edges, err := g.EdgesTo(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       val.GetGtype(),
			Expressions: val.GetExpressions(),
			Limit:       val.GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edges.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesTo.Edges = append(detail.EdgesTo.Edges, eDetail)
		}
	}
	detail.EdgesTo.Sort()
	detail.EdgesFrom.Sort()
	return detail, nil
}

func removeEdge(path *apipb.Path, paths []*apipb.Path) []*apipb.Path {
	var newPaths []*apipb.Path
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	return newPaths
}

func sortPaths(paths []*apipb.Path) {
	s := sortable.Sortable{
		LenFunc: func() int {
			return len(paths)
		},
		LessFunc: func(i, j int) bool {
			return paths[i].String() < paths[j].String()
		},
		SwapFunc: func(i, j int) {
			paths[i], paths[j] = paths[j], paths[i]
		},
	}
	s.Sort()
}

func (g *GraphStore) init(ctx context.Context) error {
	var err error
	g.initOnce.Do(func() {
		tx, err := g.db.Begin(true)
		if err != nil {
			return
		}
		defer tx.Rollback()

		// Create all the buckets
		_, err = tx.CreateBucketIfNotExists(dbNodes)
		if err != nil {
			return
		}
		_, err = tx.CreateBucketIfNotExists(dbEdges)
		if err != nil {
			return
		}
		err = tx.Commit()
		if err != nil {
			return
		}
		g.machine.Go(func(routine machine.Routine) {
			if err := g.auth.RefreshKeys(); err != nil {
				logger.Error("failed to refresh jwks", zap.Error(err))
			}
		}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
		g.machine.Go(func(routine machine.Routine) {
			g.triggerMu.Lock()
			for _, trigger := range g.triggers {
				if err := trigger.refresh(routine.Context()); err != nil {
					logger.Error("failed to refresh trigger matcher", zap.Error(err))
				}
			}
			g.triggerMu.Unlock()
		}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
		err = g.RangeEdges(ctx, apipb.Any, func(e *apipb.Edge) bool {
			g.mu.Lock()
			g.edgesFrom[e.From.String()] = append(g.edgesFrom[e.From.String()], e.Path)
			g.edgesTo[e.To.String()] = append(g.edgesTo[e.To.String()], e.Path)
			g.mu.Unlock()
			return true
		})
	})
	return err
}
