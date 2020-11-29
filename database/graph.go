package database

import (
	"context"
	"encoding/json"
	"github.com/autom8ter/graphik/cache"
	"github.com/autom8ter/graphik/flags"
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/autom8ter/graphik/logger"
	"github.com/autom8ter/graphik/vm"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"net/http"
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
	dbConnections = []byte("connections")
	dbDocs        = []byte("docs")
	// An error indicating a given key does not exist
	ErrNotFound = errors.New("not found")
)

type Graph struct {
	vm *vm.VM
	// db is the underlying handle to the db.
	db          *bbolt.DB
	jwksMu      sync.RWMutex
	jwksSet     *jwk.Set
	openID      *openIDConnect
	authorizers []cel.Program
	// The path to the Bolt database file
	path            string
	mu              sync.RWMutex
	connectionsTo   map[string][]*apipb.Path
	connectionsFrom map[string][]*apipb.Path
	machine         *machine.Machine
	closers         []func()
	closeOnce       sync.Once
	cache           *cache.Cache
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
	var closers []func()
	var programs []cel.Program
	if len(flgs.Authorizers) > 0 && flgs.Authorizers[0] != "" {
		programs, err = vMachine.Auth().Programs(flgs.Authorizers)
		if err != nil {
			return nil, err
		}
	}
	m := machine.New(ctx)
	g := &Graph{
		vm:              vMachine,
		db:              handle,
		jwksMu:          sync.RWMutex{},
		jwksSet:         nil,
		authorizers:     programs,
		path:            path,
		mu:              sync.RWMutex{},
		connectionsTo:   map[string][]*apipb.Path{},
		connectionsFrom: map[string][]*apipb.Path{},
		machine:         m,
		closers:         closers,
		closeOnce:       sync.Once{},
		cache:           cache.New(m, 1*time.Minute),
	}
	if flgs.OpenIDConnect != "" {
		resp, err := http.DefaultClient.Get(flgs.OpenIDConnect)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var openID openIDConnect
		bits, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bits, &openID); err != nil {
			return nil, err
		}
		g.openID = &openID
		set, err := jwk.Fetch(openID.JwksURI)
		if err != nil {
			return nil, err
		}
		g.jwksSet = set
	}

	err = g.db.Update(func(tx *bbolt.Tx) error {
		// Create all the buckets
		_, err = tx.CreateBucketIfNotExists(dbDocs)
		if err != nil {
			return errors.Wrap(err, "failed to create doc bucket")
		}
		_, err = tx.CreateBucketIfNotExists(dbConnections)
		if err != nil {
			return errors.Wrap(err, "failed to create connection bucket")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = g.rangeConnections(ctx, apipb.Any, func(e *apipb.Connection) bool {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.connectionsFrom[e.From.String()] = append(g.connectionsFrom[e.From.String()], e.Path)
		g.connectionsTo[e.To.String()] = append(g.connectionsTo[e.To.String()], e.Path)
		return true
	})
	if err != nil {
		return nil, err
	}
	g.machine.Go(func(routine machine.Routine) {
		if g.openID != nil {
			set, err := jwk.Fetch(g.openID.JwksURI)
			if err != nil {
				logger.Error("failed to fetch jwks", zap.Error(err))
				return
			}
			g.jwksMu.Lock()
			g.jwksSet = set
			g.jwksMu.Unlock()
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
	e, err := g.ConnectionTypes(ctx)
	if err != nil {
		return nil, err
	}
	n, err := g.DocTypes(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Schema{
		ConnectionTypes: e,
		DocTypes:        n,
	}, nil
}

func (g *Graph) Me(ctx context.Context, filter *apipb.MeFilter) (*apipb.DocDetail, error) {
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	detail := &apipb.DocDetail{
		Path:            identity.Path,
		Attributes:      identity.Attributes,
		ConnectionsFrom: &apipb.ConnectionDetails{},
		ConnectionsTo:   &apipb.ConnectionDetails{},
		Metadata:        identity.Metadata,
	}
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if filter.ConnectionsFrom != nil {
			from, err := g.ConnectionsFrom(ctx, &apipb.ConnectionFilter{
				DocPath:    identity.GetPath(),
				Gtype:      filter.GetConnectionsFrom().GetGtype(),
				Expression: filter.GetConnectionsFrom().GetExpression(),
				Limit:      filter.GetConnectionsFrom().GetLimit(),
				Sort:       filter.GetConnectionsFrom().GetSort(),
				Seek:       filter.GetConnectionsFrom().GetSeek(),
				Reverse:    filter.GetConnectionsFrom().GetReverse(),
			})
			if err != nil {
				return err
			}
			for _, f := range from.GetConnections() {
				toDoc, err := g.getDoc(ctx, tx, f.To)
				if err != nil {
					return err
				}
				fromDoc, err := g.getDoc(ctx, tx, f.From)
				if err != nil {
					return err
				}
				edetail := &apipb.ConnectionDetail{
					Path:       f.Path,
					Attributes: f.Attributes,
					Directed:   f.Directed,
					From:       fromDoc,
					To:         toDoc,
					Metadata:   f.Metadata,
				}
				detail.ConnectionsFrom.Connections = append(detail.ConnectionsFrom.Connections, edetail)
			}
		}
		if filter.ConnectionsTo != nil {
			from, err := g.ConnectionsTo(ctx, &apipb.ConnectionFilter{
				DocPath:    identity.GetPath(),
				Gtype:      filter.GetConnectionsTo().GetGtype(),
				Expression: filter.GetConnectionsTo().GetExpression(),
				Limit:      filter.GetConnectionsTo().GetLimit(),
				Sort:       filter.GetConnectionsTo().GetSort(),
				Seek:       filter.GetConnectionsTo().GetSeek(),
				Reverse:    filter.GetConnectionsTo().GetReverse(),
			})
			if err != nil {
				return err
			}
			for _, f := range from.GetConnections() {
				toDoc, err := g.getDoc(ctx, tx, f.To)
				if err != nil {
					return err
				}
				fromDoc, err := g.getDoc(ctx, tx, f.From)
				if err != nil {
					return err
				}
				edetail := &apipb.ConnectionDetail{
					Path:       f.Path,
					Attributes: f.Attributes,
					Directed:   f.Directed,
					From:       fromDoc,
					To:         toDoc,
					Metadata:   f.Metadata,
				}
				detail.ConnectionsTo.Connections = append(detail.ConnectionsTo.Connections, edetail)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return detail, nil
}

func (g *Graph) CreateDocs(ctx context.Context, constructors *apipb.DocConstructors) (*apipb.Docs, error) {
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
	var docs = &apipb.Docs{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket(dbDocs)
		for _, constructor := range constructors.GetDocs() {
			bucket := docBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
			if constructor.Path.GetGid() == "" {
				constructor.Path.Gid = uuid.New().String()
			}
			doc := &apipb.Doc{
				Path:       constructor.GetPath(),
				Attributes: constructor.GetAttributes(),
				Metadata: &apipb.Metadata{
					CreatedAt: now,
					UpdatedAt: now,
					UpdatedBy: identity.GetPath(),
				},
			}
			if bucket == nil {
				bucket, err = docBucket.CreateBucketIfNotExists([]byte(doc.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			seq, _ := bucket.NextSequence()
			doc.Metadata.Sequence = seq
			doc, err := g.setDoc(ctx, tx, doc)
			if err != nil {
				return err
			}
			docs.Docs = append(docs.Docs, doc)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var changes []*apipb.DocChange
	for _, doc := range docs.GetDocs() {
		changes = append(changes, &apipb.DocChange{
			After: doc,
		})
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:     method,
		Identity:   identity,
		Timestamp:  now,
		DocChanges: changes,
	})
	docs.Sort("")
	return docs, nil
}

func (g *Graph) CreateConnection(ctx context.Context, constructor *apipb.ConnectionConstructor) (*apipb.Connection, error) {
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	connections, err := g.CreateConnections(ctx, &apipb.ConnectionConstructors{Connections: []*apipb.ConnectionConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return connections.GetConnections()[0], nil
}

func (g *Graph) CreateConnections(ctx context.Context, constructors *apipb.ConnectionConstructors) (*apipb.Connections, error) {
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
	var connections = []*apipb.Connection{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		connectionBucket := tx.Bucket(dbConnections)
		for _, constructor := range constructors.GetConnections() {
			bucket := connectionBucket.Bucket([]byte(constructor.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = connectionBucket.CreateBucketIfNotExists([]byte(constructor.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if constructor.Path.GetGid() == "" {
				constructor.Path.Gid = uuid.New().String()
			}
			connection := &apipb.Connection{
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
			connection.Metadata.Sequence = seq
			connections = append(connections, connection)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	connectionss, err := g.setConnections(ctx, connections...)
	if err != nil {
		return nil, err
	}
	var changes []*apipb.ConnectionChange
	for _, doc := range connections {
		changes = append(changes, &apipb.ConnectionChange{
			After: doc,
		})
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:            method,
		Identity:          identity,
		Timestamp:         now,
		ConnectionChanges: changes,
	})
	connectionss.Sort("")
	return connectionss, nil
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

func (g *Graph) Subscribe(filter *apipb.ChannelFilter, server apipb.DatabaseService_SubscribeServer) error {
	var filterFunc func(msg interface{}) bool
	if filter.Expression == "" {
		filterFunc = func(msg interface{}) bool {
			_, ok := msg.(*apipb.Message)
			if !ok {
				return false
			}
			return true
		}
	} else {
		programs, err := g.vm.Message().Program(filter.Expression)
		if err != nil {
			return err
		}
		filterFunc = func(msg interface{}) bool {
			val, ok := msg.(*apipb.Message)
			if !ok {
				logger.Error("invalid message type received during subscription")
				return false
			}
			result, err := g.vm.Message().Eval(val, programs)
			if err != nil {
				logger.Error("subscription filter failure", zap.Error(err))
				return false
			}
			return result
		}
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

func (g *Graph) SubscribeChanges(filter *apipb.ExpressionFilter, server apipb.DatabaseService_SubscribeChangesServer) error {
	var (
		program cel.Program
		err     error
	)
	if filter.Expression != "" {
		program, err = g.vm.Connection().Program(filter.Expression)
		if err != nil {
			return err
		}
	}
	filterFunc := func(msg interface{}) bool {
		val, ok := msg.(*apipb.Change)
		if !ok {
			logger.Error("invalid message type received during change subscription")
			return false
		}
		if program != nil {
			result, err := g.vm.Change().Eval(val, program)
			if err != nil {
				logger.Error("subscription change failure", zap.Error(err))
				return false
			}
			return result
		}
		return true
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
	docs, err := r.AllDocs(ctx)
	if err != nil {
		return nil, err
	}
	connections, err := r.AllConnections(ctx)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Docs:        docs,
		Connections: connections,
	}, nil
}

func (r *Graph) Import(ctx context.Context, graph *apipb.Graph) (*apipb.Graph, error) {
	identity := r.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	docs, err := r.setDocs(ctx, graph.GetDocs().GetDocs()...)
	if err != nil {
		return nil, err
	}
	connections, err := r.setConnections(ctx, graph.GetConnections().GetConnections()...)
	if err != nil {
		return nil, err
	}
	return &apipb.Graph{
		Docs:        docs,
		Connections: connections,
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

func (g *Graph) GetConnection(ctx context.Context, path *apipb.Path) (*apipb.Connection, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var (
		connection *apipb.Connection
		err        error
	)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		connection, err = g.getConnection(ctx, tx, path)
		if err != nil {
			return err
		}
		return nil
	}); err != nil && err != DONE {
		return nil, err
	}
	return connection, err
}

func (n *Graph) AllDocs(ctx context.Context) (*apipb.Docs, error) {
	identity := n.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var docs []*apipb.Doc
	if err := n.rangeDocs(ctx, apipb.Any, func(doc *apipb.Doc) bool {
		docs = append(docs, doc)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Docs{
		Docs: docs,
	}
	toReturn.Sort("")
	return toReturn, nil
}

func (g *Graph) GetDoc(ctx context.Context, path *apipb.Path) (*apipb.Doc, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var (
		doc *apipb.Doc
		err error
	)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		doc, err = g.getDoc(ctx, tx, path)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return doc, nil
}

func (g *Graph) CreateDoc(ctx context.Context, constructor *apipb.DocConstructor) (*apipb.Doc, error) {
	docs, err := g.CreateDocs(ctx, &apipb.DocConstructors{Docs: []*apipb.DocConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return docs.GetDocs()[0], nil
}

func (n *Graph) PatchDoc(ctx context.Context, value *apipb.Patch) (*apipb.Doc, error) {
	identity := n.getIdentity(ctx)
	var doc *apipb.Doc
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		doc, err = n.getDoc(ctx, tx, value.GetPath())
		if err != nil {
			return err
		}
		docChange := &apipb.DocChange{
			Before: doc,
		}
		for k, v := range value.GetAttributes().GetFields() {
			doc.Attributes.GetFields()[k] = v
		}
		doc, err = n.setDoc(ctx, tx, doc)
		if err != nil {
			return err
		}
		docChange.After = doc
		n.machine.PubSub().Publish(changeChannel, &apipb.Change{
			Method:            n.getMethod(ctx),
			Identity:          identity,
			Timestamp:         doc.Metadata.UpdatedAt,
			ConnectionChanges: nil,
			DocChanges:        []*apipb.DocChange{docChange},
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return doc, err
}

func (n *Graph) PatchDocs(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Docs, error) {
	identity := n.getIdentity(ctx)
	var changes []*apipb.DocChange
	var docs []*apipb.Doc
	method := n.getMethod(ctx)
	now := timestamppb.Now()
	before, err := n.SearchDocs(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, doc := range before.GetDocs() {
		change := &apipb.DocChange{
			Before: doc,
		}
		for k, v := range patch.GetAttributes().GetFields() {
			doc.Attributes.GetFields()[k] = v
		}
		doc.GetMetadata().UpdatedAt = now
		doc.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = doc
		docs = append(docs, doc)
		changes = append(changes, change)
	}

	docss, err := n.setDocs(ctx, docs...)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:            method,
		Identity:          identity,
		Timestamp:         now,
		ConnectionChanges: nil,
		DocChanges:        changes,
	})
	return docss, nil
}

func (g *Graph) ConnectionTypes(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var types []string
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbConnections).ForEach(func(name []byte, _ []byte) error {
			types = append(types, string(name))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	sort.Strings(types)
	return types, nil
}

func (g *Graph) DocTypes(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var types []string
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbDocs).ForEach(func(name []byte, _ []byte) error {
			types = append(types, string(name))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	sort.Strings(types)
	return types, nil
}

func (g *Graph) ConnectionsFrom(ctx context.Context, filter *apipb.ConnectionFilter) (*apipb.Connections, error) {
	var (
		program cel.Program
		err     error
	)
	if filter.Expression != "" {
		program, err = g.vm.Connection().Program(filter.Expression)
		if err != nil {
			return nil, err
		}
	}
	var connections []*apipb.Connection
	var pass bool
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if err = g.rangeFrom(ctx, tx, filter.DocPath, func(connection *apipb.Connection) bool {
			if filter.Gtype != "*" {
				if connection.GetPath().GetGtype() != filter.Gtype {
					return true
				}
			}
			if program != nil {
				pass, err = g.vm.Connection().Eval(connection, program)
				if err != nil {
					return true
				}
				if pass {
					connections = append(connections, connection)
				}
			} else {
				connections = append(connections, connection)
			}
			return len(connections) < int(filter.Limit)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	toReturn := &apipb.Connections{
		Connections: connections,
	}
	toReturn.Sort("")
	return toReturn, err
}

func (n *Graph) HasDoc(ctx context.Context, path *apipb.Path) bool {
	doc, _ := n.GetDoc(ctx, path)
	return doc != nil
}

func (n *Graph) HasConnection(ctx context.Context, path *apipb.Path) bool {
	connection, _ := n.GetConnection(ctx, path)
	return connection != nil
}

func (n *Graph) SearchDocs(ctx context.Context, filter *apipb.Filter) (*apipb.Docs, error) {
	var docs []*apipb.Doc
	var program cel.Program
	var err error
	if filter.Expression != "" {
		program, err = n.vm.Doc().Program(filter.Expression)
		if err != nil {
			return nil, err
		}
	}
	seek, err := n.rangeSeekDocs(ctx, filter.Gtype, filter.GetSeek(), func(doc *apipb.Doc) bool {
		if program != nil {
			pass, err := n.vm.Doc().Eval(doc, program)
			if err != nil {
				return true
			}
			if pass {
				docs = append(docs, doc)
			}
		} else {
			docs = append(docs, doc)
		}
		return len(docs) < int(filter.Limit)
	})
	if err != nil {
		return nil, err
	}
	toReturn := &apipb.Docs{
		Docs:     docs,
		SeekNext: seek,
	}
	toReturn.Sort(filter.GetSort())
	if filter.Reverse {
		toReturn.Reverse()
	}
	return toReturn, nil
}

func (g *Graph) ConnectionsTo(ctx context.Context, filter *apipb.ConnectionFilter) (*apipb.Connections, error) {
	var (
		program cel.Program
		err     error
	)
	if filter.Expression != "" {
		program, err = g.vm.Connection().Program(filter.Expression)
		if err != nil {
			return nil, err
		}
	}
	var connections []*apipb.Connection
	var pass bool
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if err = g.rangeTo(ctx, tx, filter.DocPath, func(connection *apipb.Connection) bool {
			if filter.Gtype != "*" {
				if connection.GetPath().GetGtype() != filter.Gtype {
					return true
				}
			}
			if program != nil {
				pass, err = g.vm.Connection().Eval(connection, program)
				if err != nil {
					return true
				}
				if pass {
					connections = append(connections, connection)
				}
			} else {
				connections = append(connections, connection)
			}
			return len(connections) < int(filter.Limit)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	toReturn := &apipb.Connections{
		Connections: connections,
	}
	toReturn.Sort(filter.GetSort())
	if filter.Reverse {
		toReturn.Reverse()
	}
	return toReturn, err
}

func (n *Graph) AllConnections(ctx context.Context) (*apipb.Connections, error) {
	var connections []*apipb.Connection
	if err := n.rangeConnections(ctx, apipb.Any, func(connection *apipb.Connection) bool {
		connections = append(connections, connection)
		return true
	}); err != nil {
		return nil, err
	}
	toReturn := &apipb.Connections{
		Connections: connections,
	}
	toReturn.Sort("")
	return toReturn, nil
}

func (n *Graph) PatchConnection(ctx context.Context, value *apipb.Patch) (*apipb.Connection, error) {
	identity := n.getIdentity(ctx)
	var connection *apipb.Connection
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbConnections).Bucket([]byte(value.GetPath().Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var e apipb.Connection
		bits := bucket.Get([]byte(value.GetPath().Gid))
		if err := proto.Unmarshal(bits, &e); err != nil {
			return err
		}
		change := &apipb.ConnectionChange{
			Before: &e,
		}
		for k, v := range value.GetAttributes().GetFields() {
			connection.Attributes.GetFields()[k] = v
		}
		connection.GetMetadata().UpdatedAt = timestamppb.Now()
		connection.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = &e
		connection, err = n.setConnection(ctx, tx, connection)
		if err != nil {
			return err
		}
		n.machine.PubSub().Publish(changeChannel, &apipb.Change{
			Method:            n.getMethod(ctx),
			Identity:          identity,
			Timestamp:         connection.Metadata.UpdatedAt,
			ConnectionChanges: []*apipb.ConnectionChange{change},
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return connection, nil
}

func (n *Graph) PatchConnections(ctx context.Context, patch *apipb.PatchFilter) (*apipb.Connections, error) {
	identity := n.getIdentity(ctx)
	var changes []*apipb.ConnectionChange
	var connections []*apipb.Connection
	method := n.getMethod(ctx)
	now := timestamppb.Now()
	before, err := n.SearchConnections(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, connection := range before.GetConnections() {
		change := &apipb.ConnectionChange{
			Before: connection,
		}
		for k, v := range patch.GetAttributes().GetFields() {
			connection.Attributes.GetFields()[k] = v
		}
		connection.GetMetadata().UpdatedAt = now
		connection.GetMetadata().UpdatedBy = identity.GetPath()
		change.After = connection
		connections = append(connections, connection)
		changes = append(changes, change)
	}

	connectionss, err := n.setConnections(ctx, connections...)
	if err != nil {
		return nil, err
	}
	n.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:            method,
		Identity:          identity,
		Timestamp:         now,
		ConnectionChanges: changes,
	})
	return connectionss, nil
}

func (e *Graph) SearchConnections(ctx context.Context, filter *apipb.Filter) (*apipb.Connections, error) {
	var (
		program cel.Program
		err     error
	)
	if filter.Expression != "" {
		program, err = e.vm.Connection().Program(filter.Expression)
		if err != nil {
			return nil, err
		}
	}
	var connections []*apipb.Connection
	seek, err := e.rangeSeekConnections(ctx, filter.Gtype, filter.GetSeek(), func(connection *apipb.Connection) bool {
		if program != nil {
			pass, err := e.vm.Connection().Eval(connection, program)
			if err != nil {
				return true
			}
			if pass {
				connections = append(connections, connection)
			}
		} else {
			connections = append(connections, connection)
		}
		return len(connections) < int(filter.Limit)
	})
	if err != nil {
		return nil, err
	}
	toReturn := &apipb.Connections{
		Connections: connections,
		SeekNext:    seek,
	}
	toReturn.Sort(filter.GetSort())
	if filter.Reverse {
		toReturn.Reverse()
	}
	return toReturn, nil
}

func (g *Graph) SubGraph(ctx context.Context, filter *apipb.SubGraphFilter) (*apipb.Graph, error) {
	graph := &apipb.Graph{
		Docs:        &apipb.Docs{},
		Connections: &apipb.Connections{},
	}
	docs, err := g.SearchDocs(ctx, filter.GetDocFilter())
	if err != nil {
		return nil, err
	}
	for _, doc := range docs.GetDocs() {
		graph.Docs.Docs = append(graph.Docs.Docs, doc)
		connections, err := g.ConnectionsFrom(ctx, &apipb.ConnectionFilter{
			DocPath:    doc.Path,
			Gtype:      filter.GetConnectionFilter().GetGtype(),
			Expression: filter.GetConnectionFilter().GetExpression(),
			Limit:      filter.GetConnectionFilter().GetLimit(),
		})
		if err != nil {
			return nil, err
		}
		graph.Connections.Connections = append(graph.Connections.Connections, connections.GetConnections()...)
	}
	graph.Connections.Sort(filter.GetConnectionFilter().GetSort())
	if filter.GetConnectionFilter().GetReverse() {
		graph.GetConnections().Reverse()
	}
	graph.Docs.Sort(filter.GetDocFilter().GetSort())
	if filter.GetDocFilter().GetReverse() {
		graph.GetDocs().Reverse()
	}
	return graph, err
}

func (g *Graph) GetConnectionDetail(ctx context.Context, path *apipb.Path) (*apipb.ConnectionDetail, error) {
	var (
		err        error
		connection *apipb.Connection
		from       *apipb.Doc
		to         *apipb.Doc
	)
	if err = g.db.View(func(tx *bbolt.Tx) error {
		connection, err = g.getConnection(ctx, tx, path)
		if err != nil {
			return err
		}
		from, err = g.getDoc(ctx, tx, connection.From)
		if err != nil {
			return err
		}
		to, err = g.getDoc(ctx, tx, connection.To)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &apipb.ConnectionDetail{
		Path:       connection.GetPath(),
		Attributes: connection.GetAttributes(),
		Directed:   connection.GetDirected(),
		From:       from,
		To:         to,
		Metadata:   connection.GetMetadata(),
	}, nil
}

func (g *Graph) GetDocDetail(ctx context.Context, filter *apipb.DocDetailFilter) (*apipb.DocDetail, error) {
	detail := &apipb.DocDetail{
		Path:            filter.GetPath(),
		ConnectionsTo:   &apipb.ConnectionDetails{},
		ConnectionsFrom: &apipb.ConnectionDetails{},
	}
	var (
		err error
		doc *apipb.Doc
	)
	if err = g.db.View(func(tx *bbolt.Tx) error {
		doc, err = g.getDoc(ctx, tx, filter.GetPath())
		if err != nil {
			return err
		}
		detail.Metadata = doc.Metadata
		detail.Attributes = doc.Attributes
		if val := filter.GetFromConnections(); val != nil {
			connections, err := g.ConnectionsFrom(ctx, &apipb.ConnectionFilter{
				DocPath:    doc.Path,
				Gtype:      val.GetGtype(),
				Expression: val.GetExpression(),
				Limit:      val.GetLimit(),
				Sort:       val.GetSort(),
				Seek:       val.GetSeek(),
				Reverse:    val.GetReverse(),
			})
			if err != nil {
				return err
			}
			for _, connection := range connections.GetConnections() {
				eDetail, err := g.GetConnectionDetail(ctx, connection.GetPath())
				if err != nil {
					return err
				}
				detail.ConnectionsFrom.Connections = append(detail.ConnectionsFrom.Connections, eDetail)
			}
		}

		if val := filter.GetToConnections(); val != nil {
			connections, err := g.ConnectionsTo(ctx, &apipb.ConnectionFilter{
				DocPath:    doc.Path,
				Gtype:      val.GetGtype(),
				Expression: val.GetExpression(),
				Limit:      val.GetLimit(),
				Sort:       val.GetSort(),
				Seek:       val.GetSeek(),
				Reverse:    val.GetReverse(),
			})
			if err != nil {
				return err
			}
			for _, connection := range connections.GetConnections() {
				eDetail, err := g.GetConnectionDetail(ctx, connection.GetPath())
				if err != nil {
					return err
				}
				detail.ConnectionsTo.Connections = append(detail.ConnectionsTo.Connections, eDetail)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	detail.ConnectionsTo.Sort(filter.GetToConnections().GetSort())
	if filter.GetToConnections().Reverse {
		detail.GetConnectionsTo().Reverse()
	}
	detail.ConnectionsFrom.Sort(filter.GetFromConnections().GetSort())
	if filter.GetFromConnections().Reverse {
		detail.GetConnectionsFrom().Reverse()
	}
	return detail, err
}

func (g *Graph) DelDoc(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	var (
		n   *apipb.Doc
		err error
	)
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		n, err = g.getDoc(ctx, tx, path)
		if err != nil {
			return err
		}
		if err := g.delDoc(ctx, tx, path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:    g.getMethod(ctx),
		Identity:  g.getIdentity(ctx),
		Timestamp: timestamppb.Now(),
		DocChanges: []*apipb.DocChange{{
			Before: n,
			After:  nil,
		}},
	})
	return &empty.Empty{}, nil
}

func (g *Graph) DelDocs(ctx context.Context, filter *apipb.Filter) (*empty.Empty, error) {
	var changes []*apipb.DocChange
	before, err := g.SearchDocs(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(before.GetDocs()) == 0 {
		return nil, ErrNotFound
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, doc := range before.GetDocs() {
			if err := g.delDoc(ctx, tx, doc.GetPath()); err != nil {
				return err
			}
			changes = append(changes, &apipb.DocChange{Before: doc})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:     g.getMethod(ctx),
		Identity:   g.getIdentity(ctx),
		Timestamp:  timestamppb.Now(),
		DocChanges: changes,
	})
}

func (g *Graph) DelConnection(ctx context.Context, path *apipb.Path) (*empty.Empty, error) {
	var (
		n   *apipb.Connection
		err error
	)
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		n, err = g.getConnection(ctx, tx, path)
		if err != nil {
			return err
		}
		if err := g.delConnection(ctx, tx, path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:    g.getMethod(ctx),
		Identity:  g.getIdentity(ctx),
		Timestamp: timestamppb.Now(),
		ConnectionChanges: []*apipb.ConnectionChange{{
			Before: n,
			After:  nil,
		}},
	})
	return &empty.Empty{}, nil
}

func (g *Graph) DelConnections(ctx context.Context, filter *apipb.Filter) (*empty.Empty, error) {
	var changes []*apipb.ConnectionChange
	before, err := g.SearchConnections(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(before.GetConnections()) == 0 {
		return nil, ErrNotFound
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, doc := range before.GetConnections() {
			if err := g.delConnection(ctx, tx, doc.GetPath()); err != nil {
				return err
			}
			changes = append(changes, &apipb.ConnectionChange{Before: doc})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, g.machine.PubSub().Publish(changeChannel, &apipb.Change{
		Method:            g.getMethod(ctx),
		Identity:          g.getIdentity(ctx),
		Timestamp:         timestamppb.Now(),
		ConnectionChanges: changes,
	})
}

type openIDConnect struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}
