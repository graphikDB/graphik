package graph

import (
	"context"
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
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
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
	edgesTo     map[string][]*apipb.Path
	edgesFrom   map[string][]*apipb.Path
	machine     *machine.Machine
	triggers    []*triggerClient
	authorizers []cel.Program
	auth        *auth.Config
	closers     []func()
	closeOnce   sync.Once
}

// NewGraphStore takes a file path and returns a connected Raft backend.
func NewGraphStore(ctx context.Context, flgs *flags.Flags) (*GraphStore, error) {
	os.MkdirAll(flgs.StoragePath, 0700)
	path := filepath.Join(flgs.StoragePath, "graph.db")
	handle, err := bbolt.Open(path, dbFileMode, nil)
	if err != nil {
		return nil, err
	}
	tx, err := handle.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(dbNodes); err != nil {
		return nil, err
	}
	if _, err := tx.CreateBucketIfNotExists(dbEdges); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
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
	m := machine.New(ctx)
	m.Go(func(routine machine.Routine) {
		if err := jwks.RefreshKeys(); err != nil {
			logger.Error("failed to refresh jwks", zap.Error(err))
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(5*time.Minute))))
	return &GraphStore{
		db:          handle,
		path:        path,
		mu:          sync.RWMutex{},
		edgesTo:     map[string][]*apipb.Path{},
		edgesFrom:   map[string][]*apipb.Path{},
		machine:     m,
		triggers:    triggers,
		authorizers: programs,
		auth:        jwks,
		closers:     closers,
		closeOnce:   sync.Once{},
	}, nil
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
		EdgesTo:    map[string]*apipb.EdgeDetails{},
		EdgesFrom:  map[string]*apipb.EdgeDetails{},
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
				Cascade:    f.Cascade,
				From:       fromNode,
				To:         toNode,
				Metadata:   f.Metadata,
			}
			if _, ok := detail.EdgesFrom[f.GetPath().GetGtype()]; !ok {
				detail.EdgesFrom[f.GetPath().GetGtype()] = &apipb.EdgeDetails{}
			}
			detail.EdgesFrom[f.GetPath().GetGtype()].Edges = append(detail.EdgesFrom[f.GetPath().GetGtype()].Edges, edetail)
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
				Cascade:    f.Cascade,
				From:       fromNode,
				To:         toNode,
				Metadata:   f.Metadata,
			}
			if _, ok := detail.EdgesTo[f.GetPath().GetGtype()]; !ok {
				detail.EdgesTo[f.GetPath().GetGtype()] = &apipb.EdgeDetails{}
			}
			detail.EdgesTo[f.GetPath().GetGtype()].Edges = append(detail.EdgesTo[f.GetPath().GetGtype()].Edges, edetail)
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
	now := time.Now()
	var nodes = &apipb.Nodes{Nodes: []*apipb.Node{}}
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
					CreatedAt: now.UnixNano(),
					UpdatedAt: now.UnixNano(),
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
			bits, err := proto.Marshal(node)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(node.GetPath().GetGid()), bits); err != nil {
				return err
			}
			nodes.Nodes = append(nodes.Nodes, node)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	nodes.Sort()
	return nodes, nil
}

func (g *GraphStore) DelNode(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	identity := g.NodeContext(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	return g.DelNodes(ctx, &apipb.Paths{Paths: []*apipb.Path{path}})
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
	now := time.Now()
	var edges = &apipb.Edges{Edges: []*apipb.Edge{}}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		edgeBucket := tx.Bucket(dbEdges)
		for _, constructor := range constructors.GetEdges() {
			bucket := edgeBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
			if constructor.Path.GetGid() == "" {
				constructor.Path.Gid = uuid.New().String()
			}
			edge := &apipb.Edge{
				Path:       constructor.GetPath(),
				Attributes: constructor.GetAttributes(),
				Metadata: &apipb.Metadata{
					CreatedAt: now.UnixNano(),
					UpdatedAt: now.UnixNano(),
					UpdatedBy: identity.GetPath(),
				},
				From:    constructor.From,
				To:      constructor.To,
				Cascade: constructor.Cascade,
			}
			if bucket == nil {
				bucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			seq, _ := bucket.NextSequence()
			edge.Metadata.Sequence = seq
			bits, err := proto.Marshal(edge)
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(edge.GetPath().GetGid()), bits); err != nil {
				return err
			}
			edges.Edges = append(edges.Edges, edge)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	edges.Sort()
	return edges, nil
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
		Timestamp: time.Now().UnixNano(),
	})
}

func (g *GraphStore) Subscribe(filter *apipb.ChannelFilter, server apipb.GraphService_SubscribeServer) error {
	programs, err := vm.Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		if val, ok := msg.(apipb.Mapper); ok {
			result, err := vm.Eval(programs, val)
			if err != nil {
				logger.Error("subscription filter failure", zap.Error(err))
				return false
			}
			return result
		}
		return false
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
	now := time.Now()
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
				CreatedAt: now.UnixNano(),
				UpdatedAt: now.UnixNano(),
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
		bits, err := proto.Marshal(node)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(node.GetPath().GetGid()), bits); err != nil {
			return err
		}
		return nil
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
			bits, err := proto.Marshal(edge)
			if err != nil {
				return err
			}
			edgeBucket := tx.Bucket(dbEdges)
			bucket := edgeBucket.Bucket([]byte(edge.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if err := bucket.Put([]byte(edge.GetPath().GetGid()), bits); err != nil {
				return err
			}
			g.mu.Lock()
			defer g.mu.Unlock()
			g.edgesFrom[edge.GetFrom().String()] = append(g.edgesFrom[edge.GetFrom().String()], edge.GetPath())
			g.edgesTo[edge.GetTo().String()] = append(g.edgesTo[edge.GetTo().String()], edge.GetPath())
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

func (g *GraphStore) DelEdge(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	return g.DelEdges(ctx, &apipb.Paths{Paths: []*apipb.Path{path}})
}

func (g *GraphStore) DelEdges(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &empty.Empty{}, g.db.Update(func(tx *bbolt.Tx) error {
		for _, path := range paths.GetPaths() {
			bucket := tx.Bucket(dbEdges)
			bucket = bucket.Bucket([]byte(path.GetGtype()))
			if bucket == nil {
				return ErrNotFound
			}
			edge, err := g.GetEdge(ctx, path)
			if err != nil {
				return err
			}
			if err := bucket.Delete([]byte(path.GetGid())); err != nil {
				return err
			}
			g.mu.Lock()
			defer g.mu.Unlock()
			g.edgesFrom[edge.From.String()] = removeEdge(path, g.edgesFrom[edge.From.String()])
			g.edgesTo[edge.From.String()] = removeEdge(path, g.edgesTo[edge.From.String()])
			g.edgesFrom[edge.To.String()] = removeEdge(path, g.edgesFrom[edge.To.String()])
			g.edgesTo[edge.To.String()] = removeEdge(path, g.edgesTo[edge.To.String()])
		}
		return nil
	})
}

func (g *GraphStore) DelNodes(ctx context.Context, paths *apipb.Paths) (*empty.Empty, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &empty.Empty{}, g.db.Update(func(tx *bbolt.Tx) error {
		for _, p := range paths.GetPaths() {
			bucket := tx.Bucket(dbNodes)
			bucket = bucket.Bucket([]byte(p.GetGtype()))
			if bucket == nil {
				return ErrNotFound
			}
			return bucket.Delete([]byte(p.GetGid()))
		}
		return nil
	})
}

func (n *GraphStore) PatchNode(ctx context.Context, value *apipb.Patch) (*apipb.Node, error) {
	patch, err := n.PatchNodes(ctx, &apipb.Patches{Patches: []*apipb.Patch{value}})
	if err != nil {
		return nil, err
	}
	return patch.GetNodes()[0], nil
}

func (n *GraphStore) PatchNodes(ctx context.Context, values *apipb.Patches) (*apipb.Nodes, error) {
	identity := n.NodeContext(ctx)
	var nodes []*apipb.Node
	for _, val := range values.GetPatches() {
		node, err := n.GetNode(ctx, val.Path)
		if err != nil {
			return nil, err
		}
		for k, v := range val.GetAttributes().GetFields() {
			node.Attributes.GetFields()[k] = v
		}
		nodes = append(nodes, node)
	}
	for _, node := range nodes {
		node.GetMetadata().UpdatedAt = time.Now().UnixNano()
		node.GetMetadata().UpdatedBy = identity.GetPath()
	}

	return n.SetNodes(ctx, nodes...)
}

func (g *GraphStore) DelNodeType(ctx context.Context, typ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes)
		return bucket.DeleteBucket([]byte(typ))
	})
}

func (g *GraphStore) DelEdgeType(ctx context.Context, typ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges)
		return bucket.DeleteBucket([]byte(typ))
	})
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
	patch, err := n.PatchEdges(ctx, &apipb.Patches{Patches: []*apipb.Patch{value}})
	if err != nil {
		return nil, err
	}
	return patch.GetEdges()[0], nil
}

func (n *GraphStore) PatchEdges(ctx context.Context, values *apipb.Patches) (*apipb.Edges, error) {
	identity := n.NodeContext(ctx)
	var edges []*apipb.Edge
	for _, val := range values.GetPatches() {
		edge, err := n.GetEdge(ctx, val.Path)
		if err != nil {
			return nil, err
		}
		for k, v := range val.GetAttributes().GetFields() {
			edge.Attributes.GetFields()[k] = v
		}
		edges = append(edges, edge)
	}
	for _, edge := range edges {
		edge.GetMetadata().UpdatedAt = time.Now().UnixNano()
		edge.GetMetadata().UpdatedBy = identity.GetPath()
	}
	return n.SetEdges(ctx, edges...)
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
		Cascade:    e.GetCascade(),
		From:       from,
		To:         to,
		Metadata:   e.GetMetadata(),
	}, nil
}

func (g *GraphStore) GetNodeDetail(ctx context.Context, filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	detail := &apipb.NodeDetail{
		Path:      filter.GetPath(),
		EdgesTo:   map[string]*apipb.EdgeDetails{},
		EdgesFrom: map[string]*apipb.EdgeDetails{},
	}
	node, err := g.GetNode(ctx, filter.GetPath())
	if err != nil {
		return nil, err
	}
	detail.Metadata = node.Metadata
	detail.Attributes = node.Attributes
	if filter.GetEdgesFrom() != nil {
		edges, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdgesFrom().GetGtype(),
			Expressions: filter.GetEdgesFrom().GetExpressions(),
			Limit:       filter.GetEdgesFrom().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edges.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesFrom[edge.GetPath().GetGtype()].Edges = append(detail.EdgesFrom[edge.GetPath().GetGtype()].Edges, eDetail)
		}
	}

	if filter.GetEdgesTo() != nil {
		edges, err := g.EdgesTo(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdgesTo().GetGtype(),
			Expressions: filter.GetEdgesTo().GetExpressions(),
			Limit:       filter.GetEdgesTo().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		for _, edge := range edges.GetEdges() {
			eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
			if err != nil {
				return nil, err
			}
			detail.EdgesTo[edge.GetPath().GetGtype()].Edges = append(detail.EdgesTo[edge.GetPath().GetGtype()].Edges, eDetail)
		}
	}
	for _, d := range detail.EdgesTo {
		d.Sort()
	}
	for _, d := range detail.EdgesFrom {
		d.Sort()
	}
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
