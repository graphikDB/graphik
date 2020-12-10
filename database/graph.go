package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/autom8ter/machine"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/generic"
	"github.com/graphikDB/graphik/helpers"
	"github.com/graphikDB/graphik/logger"
	"github.com/graphikDB/graphik/vm"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
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
	"strings"
	"sync"
	"time"
)

type Graph struct {
	vm *vm.VM
	// db is the underlying handle to the db.
	db              *bbolt.DB
	jwksMu          sync.RWMutex
	jwksSet         *jwk.Set
	jwtCache        *generic.Cache
	openID          *openIDConnect
	path            string
	mu              sync.RWMutex
	connectionsTo   map[string]map[string]struct{}
	connectionsFrom map[string]map[string]struct{}
	machine         *machine.Machine
	closers         []func()
	closeOnce       sync.Once
	indexes         *generic.Cache
	authorizers     *generic.Cache
	typeValidators  *generic.Cache
	rootUsers       []string
}

// NewGraph takes a file path and returns a connected Raft backend.
func NewGraph(ctx context.Context, flgs *apipb.Flags) (*Graph, error) {
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
	m := machine.New(ctx, machine.WithMaxRoutines(100000))
	g := &Graph{
		vm:              vMachine,
		db:              handle,
		jwksMu:          sync.RWMutex{},
		jwksSet:         nil,
		path:            path,
		mu:              sync.RWMutex{},
		connectionsTo:   map[string]map[string]struct{}{},
		connectionsFrom: map[string]map[string]struct{}{},
		machine:         m,
		closers:         closers,
		closeOnce:       sync.Once{},
		jwtCache:        generic.NewCache(m, 1*time.Minute),
		indexes:         generic.NewCache(m, 1*time.Hour),
		authorizers:     generic.NewCache(m, 1*time.Hour),
		typeValidators:  generic.NewCache(m, 1*time.Hour),
		rootUsers:       flgs.RootUsers,
	}
	if flgs.OpenIdDiscovery != "" {
		resp, err := http.DefaultClient.Get(flgs.OpenIdDiscovery)
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
		_, err = tx.CreateBucketIfNotExists(dbIndexes)
		if err != nil {
			return errors.Wrap(err, "failed to create index bucket")
		}
		_, err = tx.CreateBucketIfNotExists(dbAuthorizers)
		if err != nil {
			return errors.Wrap(err, "failed to create authorizers bucket")
		}
		_, err = tx.CreateBucketIfNotExists(dbIndexDocs)
		if err != nil {
			return errors.Wrap(err, "failed to create doc/index bucket")
		}
		_, err = tx.CreateBucketIfNotExists(dbIndexConnections)
		if err != nil {
			return errors.Wrap(err, "failed to create connection/index bucket")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := g.cacheConnectionRefs(); err != nil {
		return nil, err
	}
	if err := g.cacheIndexes(); err != nil {
		return nil, err
	}
	if err := g.cacheAuthorizers(); err != nil {
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

func (g *Graph) implements() apipb.DatabaseServiceServer {
	return g
}

func (g *Graph) Ping(ctx context.Context, e *empty.Empty) (*apipb.Pong, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}

func (g *Graph) GetSchema(ctx context.Context, _ *empty.Empty) (*apipb.Schema, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	e, err := g.ConnectionTypes(ctx)
	if err != nil {
		return nil, err
	}
	n, err := g.DocTypes(ctx)
	if err != nil {
		return nil, err
	}
	var indexes []*apipb.Index
	g.rangeIndexes(func(index *index) bool {
		indexes = append(indexes, index.index)
		return true
	})
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].Name < indexes[j].Name
	})
	var authorizers []*apipb.Authorizer
	g.rangeAuthorizers(func(a *authorizer) bool {
		authorizers = append(authorizers, a.authorizer)
		return true
	})
	sort.Slice(authorizers, func(i, j int) bool {
		return authorizers[i].Name < authorizers[j].Name
	})
	var typeValidators []*apipb.TypeValidator
	g.rangeTypeValidators(func(v *typeValidator) bool {
		typeValidators = append(typeValidators, v.validator)
		return true
	})
	sort.Slice(typeValidators, func(i, j int) bool {
		ival := typeValidators[i]
		jval := typeValidators[j]
		return fmt.Sprintf("%s.%s", ival.Gtype, ival.Name) < fmt.Sprintf("%s.%s", jval.Gtype, jval.Name)
	})
	return &apipb.Schema{
		ConnectionTypes: e,
		DocTypes:        n,
		Authorizers:     &apipb.Authorizers{Authorizers: authorizers},
		Validators:      &apipb.TypeValidators{Validators: typeValidators},
		Indexes:         &apipb.Indexes{Indexes: indexes},
	}, nil
}

func (g *Graph) SetIndexes(ctx context.Context, index2 *apipb.Indexes) (*empty.Empty, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	var indexes []*apipb.Index
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, index := range index2.GetIndexes() {
			i, err := g.setIndex(ctx, tx, index)
			if err != nil {
				return err
			}
			indexes = append(indexes, i)
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, g.cacheIndexes()
}

func (g *Graph) SetAuthorizers(ctx context.Context, as *apipb.Authorizers) (*empty.Empty, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, a := range as.GetAuthorizers() {
			_, err := g.setAuthorizer(ctx, tx, a)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, g.cacheAuthorizers()
}

func (g *Graph) SetTypeValidators(ctx context.Context, as *apipb.TypeValidators) (*empty.Empty, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, v := range as.GetValidators() {
			_, err := g.setTypedValidator(ctx, tx, v)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, g.cacheTypeValidators()
}

func (g *Graph) Me(ctx context.Context, _ *empty.Empty) (*apipb.Doc, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	return g.GetDoc(ctx, user.GetRef())
}

func (g *Graph) CreateDocs(ctx context.Context, constructors *apipb.DocConstructors) (*apipb.Docs, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	method := g.getMethod(ctx)
	var docs = &apipb.Docs{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket(dbDocs)
		for _, constructor := range constructors.GetDocs() {
			bucket := docBucket.Bucket([]byte(constructor.GetRef().GetGtype()))
			if bucket == nil {
				bucket, err = docBucket.CreateBucketIfNotExists([]byte(constructor.GetRef().GetGtype()))
				if err != nil {
					return err
				}
			}
			if constructor.GetRef().Gid == "" {
				constructor.GetRef().Gid = ksuid.New().String()
			}
			path := &apipb.Ref{
				Gtype: constructor.GetRef().GetGtype(),
				Gid:   constructor.GetRef().GetGid(),
			}
			if doc, err := g.getDoc(ctx, tx, path); err == nil || doc != nil {
				return ErrAlreadyExists
			}
			doc := &apipb.Doc{
				Ref:        path,
				Attributes: constructor.GetAttributes(),
			}
			doc, err = g.setDoc(ctx, tx, doc)
			if err != nil {
				return err
			}
			if doc.GetRef().GetGid() != user.GetRef().GetGid() && doc.GetRef().GetGtype() != user.GetRef().GetGtype() {
				_, err := g.setConnection(ctx, tx, &apipb.Connection{
					Ref: &apipb.Ref{Gtype: "created", Gid: ksuid.New().String()},
					Attributes: apipb.NewStruct(map[string]interface{}{
						"method": method,
					}),
					Directed: true,
					From:     user.GetRef(),
					To:       doc.GetRef(),
				})
				if err != nil {
					return err
				}
				_, err = g.setConnection(ctx, tx, &apipb.Connection{
					Ref: &apipb.Ref{Gtype: "created_by", Gid: ksuid.New().String()},
					Attributes: apipb.NewStruct(map[string]interface{}{
						"method": method,
					}),
					Directed: true,
					To:       user.GetRef(),
					From:     doc.GetRef(),
				})
				if err != nil {
					return err
				}
			}
			docs.Docs = append(docs.Docs, doc)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	docs.Sort("")
	return docs, nil
}

func (g *Graph) CreateConnection(ctx context.Context, constructor *apipb.ConnectionConstructor) (*apipb.Connection, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	connections, err := g.CreateConnections(ctx, &apipb.ConnectionConstructors{Connections: []*apipb.ConnectionConstructor{constructor}})
	if err != nil {
		return nil, err
	}
	return connections.GetConnections()[0], nil
}

func (g *Graph) CreateConnections(ctx context.Context, constructors *apipb.ConnectionConstructors) (*apipb.Connections, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	var err error
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var connections = &apipb.Connections{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		connectionBucket := tx.Bucket(dbConnections)
		for _, constructor := range constructors.GetConnections() {
			bucket := connectionBucket.Bucket([]byte(constructor.GetRef().GetGtype()))
			if bucket == nil {
				bucket, err = connectionBucket.CreateBucketIfNotExists([]byte(constructor.GetRef().GetGtype()))
				if err != nil {
					return err
				}
			}
			if constructor.GetRef().Gid == "" {
				constructor.GetRef().Gid = ksuid.New().String()
			}
			path := &apipb.Ref{
				Gtype: constructor.GetRef().GetGtype(),
				Gid:   constructor.GetRef().GetGid(),
			}
			if conn, err := g.getConnection(ctx, tx, path); err == nil || conn != nil {
				return ErrAlreadyExists
			}
			connection := &apipb.Connection{
				Ref:        path,
				Attributes: constructor.GetAttributes(),
				Directed:   constructor.Directed,
				From:       constructor.GetFrom(),
				To:         constructor.GetTo(),
			}
			connection, err = g.setConnection(ctx, tx, connection)
			if err != nil {
				return err
			}
			connections.Connections = append(connections.Connections, connection)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	connections.Sort("")
	return connections, nil
}

func (g *Graph) Publish(ctx context.Context, message *apipb.OutboundMessage) (*empty.Empty, error) {
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	return &empty.Empty{}, g.machine.PubSub().Publish(message.Channel, &apipb.Message{
		Channel:   message.Channel,
		Data:      message.Data,
		Sender:    user.GetRef(),
		Timestamp: timestamppb.Now(),
	})
}

func (g *Graph) Subscribe(filter *apipb.ChanFilter, server apipb.DatabaseService_SubscribeServer) error {
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

func (g *Graph) GetConnection(ctx context.Context, path *apipb.Ref) (*apipb.Connection, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
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
	user := n.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
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

func (g *Graph) GetDoc(ctx context.Context, path *apipb.Ref) (*apipb.Doc, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	user := g.getIdentity(ctx)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
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

func (n *Graph) EditDoc(ctx context.Context, value *apipb.Edit) (*apipb.Doc, error) {
	user := n.getIdentity(ctx)
	var doc *apipb.Doc
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		doc, err = n.getDoc(ctx, tx, value.GetRef())
		if err != nil {
			return err
		}
		for k, v := range value.GetAttributes().GetFields() {
			doc.Attributes.GetFields()[k] = v
		}
		doc, err = n.setDoc(ctx, tx, doc)
		if err != nil {
			return err
		}
		if doc.GetRef().GetGid() != user.GetRef().GetGid() && doc.GetRef().GetGtype() != user.GetRef().GetGtype() {
			id := helpers.Hash([]byte(fmt.Sprintf("%s-%s", user.GetRef().String(), doc.GetRef().String())))
			editedRef := &apipb.Ref{Gid: id, Gtype: "edited"}
			if !n.hasConnectionFrom(user.GetRef(), editedRef) {
				_, err := n.setConnection(ctx, tx, &apipb.Connection{
					Ref:        editedRef,
					Attributes: apipb.NewStruct(map[string]interface{}{}),
					Directed:   true,
					From:       user.GetRef(),
					To:         doc.GetRef(),
				})
				if err != nil {
					return err
				}
			}
			editedByRef := &apipb.Ref{Gtype: "edited_by", Gid: id}
			if !n.hasConnectionFrom(doc.GetRef(), editedByRef) {
				_, err := n.setConnection(ctx, tx, &apipb.Connection{
					Ref:        editedByRef,
					Attributes: apipb.NewStruct(map[string]interface{}{}),
					Directed:   true,
					To:         user.GetRef(),
					From:       doc.GetRef(),
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return doc, err
}

func (n *Graph) EditDocs(ctx context.Context, patch *apipb.EFilter) (*apipb.Docs, error) {
	var docs []*apipb.Doc
	before, err := n.SearchDocs(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	for _, doc := range before.GetDocs() {
		for k, v := range patch.GetAttributes().GetFields() {
			doc.Attributes.GetFields()[k] = v
		}
		docs = append(docs, doc)
	}

	docss, err := n.setDocs(ctx, docs...)
	if err != nil {
		return nil, err
	}
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

func (g *Graph) ConnectionsFrom(ctx context.Context, filter *apipb.CFilter) (*apipb.Connections, error) {
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
		if err = g.rangeFrom(ctx, tx, filter.GetDocRef(), func(connection *apipb.Connection) bool {
			if filter.Gtype != "*" {
				if connection.GetRef().GetGtype() != filter.Gtype {
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
	seek, err := n.rangeSeekDocs(ctx, filter.Gtype, filter.GetSeek(), filter.GetIndex(), filter.GetReverse(), func(doc *apipb.Doc) bool {
		if program != nil {
			pass, err := n.vm.Doc().Eval(doc, program)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("search docs failure", zap.Error(err))
				}
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
	return toReturn, nil
}

func (n *Graph) AggregateDocs(ctx context.Context, filter *apipb.AggFilter) (*apipb.Number, error) {
	docs, err := n.SearchDocs(ctx, filter.GetFilter())
	if err != nil {
		return nil, err
	}
	return &apipb.Number{Value: docs.Aggregate(filter.GetAggregate(), filter.GetField())}, nil
}

func (n *Graph) AggregateConnections(ctx context.Context, filter *apipb.AggFilter) (*apipb.Number, error) {
	connections, err := n.SearchConnections(ctx, filter.GetFilter())
	if err != nil {
		return nil, err
	}
	return &apipb.Number{Value: connections.Aggregate(filter.GetAggregate(), filter.GetField())}, nil
}

func (n *Graph) Traverse(ctx context.Context, filter *apipb.TFilter) (*apipb.Traversals, error) {
	dfs, err := n.newTraversal(filter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := n.db.View(func(tx *bbolt.Tx) error {
		return dfs.Walk(ctx, tx)
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return dfs.traversals, nil
}

func (g *Graph) ConnectionsTo(ctx context.Context, filter *apipb.CFilter) (*apipb.Connections, error) {
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
		if err = g.rangeTo(ctx, tx, filter.GetDocRef(), func(connection *apipb.Connection) bool {
			if filter.Gtype != "*" {
				if connection.GetRef().GetGtype() != filter.Gtype {
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

func (n *Graph) EditConnection(ctx context.Context, value *apipb.Edit) (*apipb.Connection, error) {
	var connection *apipb.Connection
	var err error
	if err = n.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbConnections).Bucket([]byte(value.GetRef().Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var e apipb.Connection
		bits := bucket.Get([]byte(value.GetRef().Gid))
		if err := proto.Unmarshal(bits, &e); err != nil {
			return err
		}
		for k, v := range value.GetAttributes().GetFields() {
			connection.Attributes.GetFields()[k] = v
		}
		connection, err = n.setConnection(ctx, tx, connection)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return connection, nil
}

func (n *Graph) EditConnections(ctx context.Context, patch *apipb.EFilter) (*apipb.Connections, error) {
	var connections = &apipb.Connections{}
	before, err := n.SearchConnections(ctx, patch.GetFilter())
	if err != nil {
		return nil, err
	}
	if err := n.db.Update(func(tx *bbolt.Tx) error {
		for _, connection := range before.GetConnections() {
			for k, v := range patch.GetAttributes().GetFields() {
				connection.Attributes.GetFields()[k] = v
			}
			connection, err := n.setConnection(ctx, tx, connection)
			if err != nil {
				return err
			}
			connections.Connections = append(connections.Connections, connection)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return connections, nil
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
	seek, err := e.rangeSeekConnections(ctx, filter.Gtype, filter.GetSeek(), filter.GetIndex(), filter.GetReverse(), func(connection *apipb.Connection) bool {
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
	return toReturn, nil
}

func (g *Graph) DelDoc(ctx context.Context, path *apipb.Ref) (*empty.Empty, error) {
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		_, err := g.getDoc(ctx, tx, path)
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
	return &empty.Empty{}, nil
}

func (g *Graph) DelDocs(ctx context.Context, filter *apipb.Filter) (*empty.Empty, error) {
	before, err := g.SearchDocs(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(before.GetDocs()) == 0 {
		return nil, ErrNotFound
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, doc := range before.GetDocs() {
			if err := g.delDoc(ctx, tx, doc.GetRef()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (g *Graph) DelConnection(ctx context.Context, path *apipb.Ref) (*empty.Empty, error) {
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		if err := g.delConnection(ctx, tx, path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (g *Graph) DelConnections(ctx context.Context, filter *apipb.Filter) (*empty.Empty, error) {
	before, err := g.SearchConnections(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(before.GetConnections()) == 0 {
		return nil, ErrNotFound
	}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, doc := range before.GetConnections() {
			if err := g.delConnection(ctx, tx, doc.GetRef()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (g *Graph) PushDocConstructors(server apipb.DatabaseService_PushDocConstructorsServer) error {
	ctx, cancel := context.WithCancel(server.Context())
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			val, err := server.Recv()
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			resp, err := g.CreateDoc(ctx, val)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := server.Send(resp); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (g *Graph) PushConnectionConstructors(server apipb.DatabaseService_PushConnectionConstructorsServer) error {
	ctx, cancel := context.WithCancel(context.WithValue(server.Context(), importOverrideCtxKey, true))
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			val, err := server.Recv()
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			resp, err := g.CreateConnection(ctx, val)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if err := server.Send(resp); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}

func (g *Graph) SeedDocs(server apipb.DatabaseService_SeedDocsServer) error {
	ctx, cancel := context.WithCancel(context.WithValue(server.Context(), importOverrideCtxKey, true))
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := server.Recv()
			if err != nil {
				return err
			}
			if err := g.db.Update(func(tx *bbolt.Tx) error {
				_, err := g.setDoc(ctx, tx, msg)
				return err
			}); err != nil {
				return err
			}
		}
	}
}

func (g *Graph) SeedConnections(server apipb.DatabaseService_SeedConnectionsServer) error {
	ctx, cancel := context.WithCancel(server.Context())
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := server.Recv()
			if err != nil {
				return err
			}
			if err := g.db.Update(func(tx *bbolt.Tx) error {
				_, err := g.setConnection(ctx, tx, msg)
				return err
			}); err != nil {
				return err
			}
		}
	}
}

func (g *Graph) SearchAndConnect(ctx context.Context, filter *apipb.SConnectFilter) (*apipb.Connections, error) {
	docs, err := g.SearchDocs(ctx, filter.GetFilter())
	if err != nil {
		return nil, err
	}
	var connections []*apipb.ConnectionConstructor
	for _, doc := range docs.GetDocs() {
		connections = append(connections, &apipb.ConnectionConstructor{
			Ref: &apipb.RefConstructor{
				Gtype: filter.GetGtype(),
			},
			Attributes: filter.GetAttributes(),
			Directed:   filter.GetDirected(),
			From:       filter.GetFrom(),
			To:         doc.GetRef(),
		})
	}
	return g.CreateConnections(ctx, &apipb.ConnectionConstructors{Connections: connections})
}

func (g *Graph) ExistsDoc(ctx context.Context, has *apipb.ExistsFilter) (*apipb.Boolean, error) {
	program, err := g.vm.Doc().Program(has.GetExpression())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var (
		hasErr error
		res    bool
	)

	_, err = g.rangeSeekDocs(ctx, has.GetGtype(), has.GetSeek(), has.GetIndex(), has.GetReverse(), func(n *apipb.Doc) bool {
		res, hasErr = g.vm.Doc().Eval(n, program)
		if hasErr != nil {
			return false
		}
		if res {
			return false
		}
		return true
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if hasErr != nil {
		return nil, status.Error(codes.InvalidArgument, hasErr.Error())
	}
	return &apipb.Boolean{Value: res}, nil
}

func (g *Graph) ExistsConnection(ctx context.Context, has *apipb.ExistsFilter) (*apipb.Boolean, error) {
	program, err := g.vm.Connection().Program(has.GetExpression())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var (
		hasErr error
		res    bool
	)
	_, err = g.rangeSeekConnections(ctx, has.GetGtype(), has.GetSeek(), has.GetIndex(), has.GetReverse(), func(n *apipb.Connection) bool {
		res, hasErr = g.vm.Connection().Eval(n, program)
		if hasErr != nil {
			return false
		}
		if res {
			return false
		}
		return true
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if hasErr != nil {
		return nil, status.Error(codes.InvalidArgument, hasErr.Error())
	}
	return &apipb.Boolean{Value: res}, nil
}

func (g *Graph) HasDoc(ctx context.Context, ref *apipb.Ref) (*apipb.Boolean, error) {
	doc, err := g.GetDoc(ctx, ref)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if doc != nil {
		return &apipb.Boolean{Value: true}, nil
	}
	return &apipb.Boolean{Value: false}, nil
}

func (g *Graph) HasConnection(ctx context.Context, ref *apipb.Ref) (*apipb.Boolean, error) {
	c, err := g.GetConnection(ctx, ref)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if c != nil {
		return &apipb.Boolean{Value: true}, nil
	}
	return &apipb.Boolean{Value: false}, nil
}
