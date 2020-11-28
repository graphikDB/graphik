package graph

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/sortable"
	"github.com/autom8ter/graphik/vm"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/google/uuid"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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

type Graph struct {
	vm *vm.VM
	// db is the underlying handle to the db.
	db          *bbolt.DB
	jwksSet     *jwk.Set
	jwksSource  string
	authorizers []cel.Program
	// The path to the Bolt database file
	path      string
	mu        sync.RWMutex
	triggerMu sync.RWMutex
	edgesTo   map[string][]*apipb.Path
	edgesFrom map[string][]*apipb.Path
	machine   *machine.Machine
	triggers  []*triggerClient
	closers   []func()
	closeOnce sync.Once
}

// NewGraph takes a file path and returns a connected Raft backend.
func NewGraph(ctx context.Context, flgs *flags.Flags) (*Graph, error) {
	os.MkdirAll(flgs.StoragePath, 0700)
	path := filepath.Join(flgs.StoragePath, "graph.db")
	handle, err := bbolt.Open(path, dbFileMode, nil)
	if err != nil {
		return nil, err
	}
	vMachine, err := vm.NewVM()
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
		programs, err := vMachine.Change().Programs(matchers.GetExpressions())
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile trigger matchers")
		}
		triggers = append(triggers, &triggerClient{
			TriggerServiceClient: trig,
			matchers:             programs,
			vm:                   vMachine.Change(),
		})
		closers = append(closers, func() {
			conn.Close()
		})
	}
	var programs []cel.Program
	if len(flgs.Authorizers) > 0 && flgs.Authorizers[0] != "" {
		programs, err = vMachine.Auth().Programs(flgs.Authorizers)
		if err != nil {
			return nil, err
		}
	}
	g := &Graph{
		vm:          vMachine,
		db:          handle,
		jwksSet:     nil,
		jwksSource:  flgs.JWKS,
		authorizers: programs,
		path:        path,
		mu:          sync.RWMutex{},
		triggerMu:   sync.RWMutex{},
		edgesTo:     map[string][]*apipb.Path{},
		edgesFrom:   map[string][]*apipb.Path{},
		machine:     machine.New(ctx),
		triggers:    triggers,
		closers:     closers,
		closeOnce:   sync.Once{},
	}
	if flgs.JWKS != "" {
		set, err := jwk.Fetch(flgs.JWKS)
		if err != nil {
			return nil, err
		}
		g.jwksSet = set
	}
	err = g.db.Update(func(tx *bbolt.Tx) error {
		// Create all the buckets
		_, err = tx.CreateBucketIfNotExists(dbNodes)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(dbEdges)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = g.RangeEdges(ctx, apipb.Any, func(e *apipb.Edge) bool {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.edgesFrom[e.From.String()] = append(g.edgesFrom[e.From.String()], e.Path)
		g.edgesTo[e.To.String()] = append(g.edgesTo[e.To.String()], e.Path)
		return true
	})
	if err != nil {
		return nil, err
	}
	g.machine.Go(func(routine machine.Routine) {
		g.triggerMu.Lock()
		defer g.triggerMu.Unlock()
		if g.jwksSource != "" {
			set, err := jwk.Fetch(g.jwksSource)
			if err != nil {
				logger.Error("failed to fetch jwks", zap.Error(err))
				return
			}
			g.jwksSet = set
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
	g.machine.Go(func(routine machine.Routine) {
		g.triggerMu.Lock()
		defer g.triggerMu.Unlock()
		for _, trigger := range g.triggers {
			if err := trigger.refresh(routine.Context()); err != nil {
				logger.Error("failed to refresh trigger matcher", zap.Error(err))
			}
		}
	}, machine.GoWithMiddlewares(machine.Cron(time.NewTicker(1*time.Minute))))
	return g, nil
}

func (g *Graph) Ping(ctx context.Context, e *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (g *Graph) GetSchema(ctx context.Context, _ *empty.Empty) (*apipb.Schema, error) {
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

func (g *Graph) Me(ctx context.Context, filter *apipb.MeFilter) (*apipb.NodeDetail, error) {
	identity := g.getIdentity(ctx)
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
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if filter.EdgesFrom != nil {
			from, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
				NodePath:    identity.GetPath(),
				Gtype:       filter.GetEdgesFrom().GetGtype(),
				Expressions: filter.GetEdgesFrom().GetExpressions(),
				Limit:       filter.GetEdgesFrom().GetLimit(),
			})
			if err != nil {
				return err
			}
			for _, f := range from.GetEdges() {
				toNode, err := g.getNode(ctx, tx, f.To)
				if err != nil {
					return err
				}
				fromNode, err := g.getNode(ctx, tx, f.From)
				if err != nil {
					return err
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
				return err
			}
			for _, f := range from.GetEdges() {
				toNode, err := g.getNode(ctx, tx, f.To)
				if err != nil {
					return err
				}
				fromNode, err := g.getNode(ctx, tx, f.From)
				if err != nil {
					return err
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
		return nil
	}); err != nil {
		return nil, err
	}
	return detail, nil
}

func (g *Graph) CreateNodes(ctx context.Context, constructors *apipb.NodeConstructors) (*apipb.Nodes, error) {
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := timestamppb.Now()
	method := g.getMethod(ctx)
	var nodes = &apipb.Nodes{}
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
			node, err := g.setNode(ctx, tx, node)
			if err != nil {
				return err
			}
			nodes.Nodes = append(nodes.Nodes, node)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var changes []*apipb.NodeChange
	for _, node := range nodes.GetNodes() {
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
	nodes.Sort()
	return nodes, nil
}

func (g *Graph) CreateEdge(ctx context.Context, constructor *apipb.EdgeConstructor) (*apipb.Edge, error) {
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	edges, err := g.CreateEdges(ctx, &apipb.EdgeConstructors{Edges: []*apipb.EdgeConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return edges.GetEdges()[0], nil
}

func (g *Graph) CreateEdges(ctx context.Context, constructors *apipb.EdgeConstructors) (*apipb.Edges, error) {
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := timestamppb.Now()
	method := g.getMethod(ctx)
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
	edgess, err := g.setEdges(ctx, edges...)
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

func (g *Graph) Publish(ctx context.Context, message *apipb.OutboundMessage) (*empty.Empty, error) {
	identity := g.getIdentity(ctx)
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

func (g *Graph) Subscribe(filter *apipb.ChannelFilter, server apipb.GraphService_SubscribeServer) error {
	programs, err := g.vm.Message().Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		val, ok := msg.(*apipb.Message)
		if !ok {
			logger.Error("invalid message type received during subscription")
			return false
		}
		result, err := g.vm.Message().Eval(programs, val)
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

func (g *Graph) SubscribeChanges(filter *apipb.ExpressionFilter, server apipb.GraphService_SubscribeChangesServer) error {
	programs, err := g.vm.Change().Programs(filter.Expressions)
	if err != nil {
		return err
	}
	filterFunc := func(msg interface{}) bool {
		val, ok := msg.(*apipb.Change)
		if !ok {
			logger.Error("invalid message type received during change subscription")
			return false
		}
		result, err := g.vm.Change().Eval(programs, val)
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

func (r *Graph) Export(ctx context.Context, _ *empty.Empty) (*apipb.Graph, error) {
	identity := r.getIdentity(ctx)
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

func (r *Graph) Import(ctx context.Context, graph *apipb.Graph) (*apipb.Graph, error) {
	identity := r.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	nodes, err := r.setNodes(ctx, graph.GetNodes().GetNodes()...)
	if err != nil {
		return nil, err
	}
	edges, err := r.setEdges(ctx, graph.GetEdges().GetEdges()...)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (g *Graph) Shutdown(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	go g.Close()
	return &empty.Empty{}, nil
}

// Close is used to gracefully close the Database.
func (b *Graph) Close() {
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

func (g *Graph) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var (
		edge *apipb.Edge
		err  error
	)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		edge, err = g.getEdge(ctx, tx, path)
		if err != nil {
			return err
		}
		return nil
	}); err != nil && err != DONE {
		return nil, err
	}
	return edge, err
}

func (g *Graph) RangeSeekEdges(ctx context.Context, gType, seek string, fn func(e *apipb.Edge) bool) error {
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

func (n *Graph) AllNodes(ctx context.Context) (*apipb.Nodes, error) {
	identity := n.getIdentity(ctx)
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

func (g *Graph) RangeEdges(ctx context.Context, gType string, fn func(n *apipb.Edge) bool) error {
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

func (g *Graph) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var (
		node *apipb.Node
		err  error
	)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		node, err = g.getNode(ctx, tx, path)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (g *Graph) RangeNodes(ctx context.Context, gType string, fn func(n *apipb.Node) bool) error {
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

func (g *Graph) createIdentity(ctx context.Context, constructor *apipb.NodeConstructor) (*apipb.Node, error) {
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

		if constructor.Path.GetGid() == "" {
			constructor.Path.Gid = uuid.New().String()
		}
		nodeBucket := tx.Bucket(dbNodes)
		bucket := nodeBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
		if bucket == nil {
			bucket, err = nodeBucket.CreateBucket([]byte(node.GetPath().GetGtype()))
			if err != nil {
				return err
			}
		}
		seq, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		node.Metadata.Sequence = seq

		node, err = g.setNode(ctx, tx, node)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (g *Graph) CreateNode(ctx context.Context, constructor *apipb.NodeConstructor) (*apipb.Node, error) {
	nodes, err := g.CreateNodes(ctx, &apipb.NodeConstructors{Nodes: []*apipb.NodeConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return nodes.GetNodes()[0], nil
}

func (n *Graph) PatchNode(ctx context.Context, value *apipb.Patch) (*apipb.Node, error) {
	identity := n.getIdentity(ctx)
	var node *apipb.Node
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		node, err = n.getNode(ctx, tx, value.GetPath())
		if err != nil {
			return err
		}
		change := &apipb.NodeChange{
			Before: node,
		}
		for k, v := range value.GetAttributes().GetFields() {
			node.Attributes.GetFields()[k] = v
		}
		node, err = n.setNode(ctx, tx, node)
		if err != nil {
			return err
		}
		change.After = node
		n.machine.PubSub().Publish(changeChannel, &apipb.Change{
			Method:      n.getMethod(ctx),
			Identity:    identity,
			Timestamp:   node.Metadata.UpdatedAt,
			EdgeChanges: nil,
			NodeChanges: []*apipb.NodeChange{change},
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return node, err
}

func (n *Graph) PatchNodes(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Nodes, error) {
	identity := n.getIdentity(ctx)
	var changes []*apipb.NodeChange
	var nodes []*apipb.Node
	method := n.getMethod(ctx)
	now := timestamppb.Now()
	before, err := n.SearchNodes(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, node := range before.GetNodes() {
		change := &apipb.NodeChange{
			Before: node,
		}
		for k, v := range patch.GetAttributes().GetFields() {
			node.Attributes.GetFields()[k] = v
		}
		node.GetMetadata().UpdatedAt = now
		node.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = node
		nodes = append(nodes, node)
		changes = append(changes, change)
	}

	nodess, err := n.setNodes(ctx, nodes...)
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

func (g *Graph) EdgeTypes(ctx context.Context) ([]string, error) {
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

func (g *Graph) NodeTypes(ctx context.Context) ([]string, error) {
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

func (g *Graph) rangeFrom(ctx context.Context, tx *bbolt.Tx, nodePath *apipb.Path, fn func(e *apipb.Edge) bool) error {
	g.mu.RLock()
	paths := g.edgesFrom[nodePath.String()]
	g.mu.RUnlock()
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := tx.Bucket(dbEdges).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var edge apipb.Edge
		bits := bucket.Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, &edge); err != nil {
			return err
		}
		if !fn(&edge) {
			return nil
		}
	}
	return nil
}

func (g *Graph) rangeTo(ctx context.Context, tx *bbolt.Tx, nodePath *apipb.Path, fn func(e *apipb.Edge) bool) error {
	g.mu.RLock()
	paths := g.edgesTo[nodePath.String()]
	g.mu.RUnlock()
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := tx.Bucket(dbEdges).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var edge apipb.Edge
		bits := bucket.Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, &edge); err != nil {
			return err
		}
		if !fn(&edge) {
			return nil
		}
	}
	return nil
}

func (g *Graph) EdgesFrom(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := g.vm.Edge().Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if err = g.rangeFrom(ctx, tx, filter.NodePath, func(edge *apipb.Edge) bool {
			if filter.Gtype != "*" {
				if edge.GetPath().GetGtype() != filter.Gtype {
					return true
				}
			}

			pass, err = g.vm.Edge().Eval(programs, edge)
			if err != nil {
				return true
			}
			if pass {
				edges = append(edges, edge)
			}
			return len(edges) < int(filter.Limit)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, err
}

func (n *Graph) HasNode(ctx context.Context, path *apipb.Path) bool {
	node, _ := n.GetNode(ctx, path)
	return node != nil
}

func (n *Graph) HasEdge(ctx context.Context, path *apipb.Path) bool {
	edge, _ := n.GetEdge(ctx, path)
	return edge != nil
}

func (n *Graph) SearchNodes(ctx context.Context, filter *apipb.Filter) (*apipb.Nodes, error) {
	var nodes []*apipb.Node
	programs, err := n.vm.Node().Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	if err := n.RangeNodes(ctx, filter.Gtype, func(node *apipb.Node) bool {
		pass, err := n.vm.Node().Eval(programs, node)
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

func (n *Graph) FilterNode(ctx context.Context, nodeType string, filter func(node *apipb.Node) bool) (*apipb.Nodes, error) {
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

func (g *Graph) EdgesTo(ctx context.Context, filter *apipb.EdgeFilter) (*apipb.Edges, error) {
	programs, err := g.vm.Edge().Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	var pass bool
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if err = g.rangeTo(ctx, tx, filter.NodePath, func(edge *apipb.Edge) bool {
			if filter.Gtype != "*" {
				if edge.GetPath().GetGtype() != filter.Gtype {
					return true
				}
			}

			pass, err = g.vm.Edge().Eval(programs, edge)
			if err != nil {
				return true
			}
			if pass {
				edges = append(edges, edge)
			}
			return len(edges) < int(filter.Limit)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	toReturn := &apipb.Edges{
		Edges: edges,
	}
	toReturn.Sort()
	return toReturn, err
}

func (n *Graph) AllEdges(ctx context.Context) (*apipb.Edges, error) {
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

func (n *Graph) PatchEdge(ctx context.Context, value *apipb.Patch) (*apipb.Edge, error) {
	identity := n.getIdentity(ctx)
	var edge *apipb.Edge
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges).Bucket([]byte(value.GetPath().Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var e apipb.Edge
		bits := bucket.Get([]byte(value.GetPath().Gid))
		if err := proto.Unmarshal(bits, &e); err != nil {
			return err
		}
		change := &apipb.EdgeChange{
			Before: &e,
		}
		for k, v := range value.GetAttributes().GetFields() {
			edge.Attributes.GetFields()[k] = v
		}
		edge.GetMetadata().UpdatedAt = timestamppb.Now()
		edge.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = &e
		edge, err = n.setEdge(ctx, tx, edge)
		if err != nil {
			return err
		}
		n.machine.PubSub().Publish(changeChannel, &apipb.Change{
			Method:      n.getMethod(ctx),
			Identity:    identity,
			Timestamp:   edge.Metadata.UpdatedAt,
			EdgeChanges: []*apipb.EdgeChange{change},
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return edge, nil
}

func (n *Graph) PatchEdges(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Edges, error) {
	identity := n.getIdentity(ctx)
	var changes []*apipb.EdgeChange
	var edges []*apipb.Edge
	method := n.getMethod(ctx)
	now := timestamppb.Now()
	before, err := n.SearchEdges(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, edge := range before.GetEdges() {
		change := &apipb.EdgeChange{
			Before: edge,
		}
		for k, v := range patch.GetAttributes().GetFields() {
			edge.Attributes.GetFields()[k] = v
		}
		edge.GetMetadata().UpdatedAt = now
		edge.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = edge
		edges = append(edges, edge)
		changes = append(changes, change)
	}

	edgess, err := n.setEdges(ctx, edges...)
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

func (e *Graph) SearchEdges(ctx context.Context, filter *apipb.Filter) (*apipb.Edges, error) {
	programs, err := e.vm.Edge().Programs(filter.Expressions)
	if err != nil {
		return nil, err
	}
	var edges []*apipb.Edge
	if err := e.RangeEdges(ctx, filter.Gtype, func(edge *apipb.Edge) bool {
		pass, err := e.vm.Edge().Eval(programs, edge)
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

func (g *Graph) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	graph := &apipb.Graph{
		Nodes: &apipb.Nodes{},
		Edges: &apipb.Edges{},
	}
	nodes, err := g.SearchNodes(ctx, filter.GetNodeFilter())
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.GetNodes() {
		graph.Nodes.Nodes = append(graph.Nodes.Nodes, node)
		edges, err := g.EdgesFrom(ctx, &apipb.EdgeFilter{
			NodePath:    node.Path,
			Gtype:       filter.GetEdgeFilter().GetGtype(),
			Expressions: filter.GetEdgeFilter().GetExpressions(),
			Limit:       filter.GetEdgeFilter().GetLimit(),
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

func (g *Graph) GetEdgeDetail(ctx context.Context, path *apipb.Path) (*apipb.EdgeDetail, error) {
	var (
		err  error
		edge *apipb.Edge
		from *apipb.Node
		to   *apipb.Node
	)
	if err = g.db.View(func(tx *bbolt.Tx) error {
		edge, err = g.getEdge(ctx, tx, path)
		if err != nil {
			return err
		}
		from, err = g.getNode(ctx, tx, edge.From)
		if err != nil {
			return err
		}
		to, err = g.getNode(ctx, tx, edge.To)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &apipb.EdgeDetail{
		Path:       edge.GetPath(),
		Attributes: edge.GetAttributes(),
		Directed:   edge.GetDirected(),
		From:       from,
		To:         to,
		Metadata:   edge.GetMetadata(),
	}, nil
}

func (g *Graph) GetNodeDetail(ctx context.Context, filter *apipb.NodeDetailFilter) (*apipb.NodeDetail, error) {
	detail := &apipb.NodeDetail{
		Path:      filter.GetPath(),
		EdgesTo:   &apipb.EdgeDetails{},
		EdgesFrom: &apipb.EdgeDetails{},
	}
	var (
		err  error
		node *apipb.Node
	)
	if err = g.db.View(func(tx *bbolt.Tx) error {
		node, err = g.getNode(ctx, tx, filter.GetPath())
		if err != nil {
			return err
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
				return err
			}
			for _, edge := range edges.GetEdges() {
				eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
				if err != nil {
					return err
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
				return err
			}
			for _, edge := range edges.GetEdges() {
				eDetail, err := g.GetEdgeDetail(ctx, edge.GetPath())
				if err != nil {
					return err
				}
				detail.EdgesTo.Edges = append(detail.EdgesTo.Edges, eDetail)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	detail.EdgesTo.Sort()
	detail.EdgesFrom.Sort()
	return detail, err
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
