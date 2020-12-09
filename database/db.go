package database

import (
	"context"
	"fmt"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/graphik/logger"
	"github.com/google/cel-go/cel"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"sort"
	"strings"
)

type index struct {
	index   *apipb.Index
	program cel.Program
}

type authorizer struct {
	authorizer *apipb.Authorizer
	program    cel.Program
}

type typeValidator struct {
	validator *apipb.TypeValidator
	program   cel.Program
}

func (g *Graph) rangeIndexes(fn func(index *index) bool) {
	g.indexes.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		index := value.(*index)
		return fn(index)
	})
}

func (g *Graph) cacheConnectionRefs() error {
	return g.rangeConnections(context.Background(), apipb.Any, func(e *apipb.Connection) bool {
		g.mu.Lock()
		defer g.mu.Unlock()
		if g.connectionsFrom[e.From.String()] == nil {
			g.connectionsFrom[e.From.String()] = map[string]struct{}{}
		}
		if g.connectionsTo[e.To.String()] == nil {
			g.connectionsTo[e.To.String()] = map[string]struct{}{}
		}
		refstr := refString(e.GetRef())
		g.connectionsFrom[e.From.String()][refstr] = struct{}{}
		g.connectionsTo[e.To.String()][refstr] = struct{}{}
		return true
	})
}

func (g *Graph) cacheIndexes() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbIndexes).ForEach(func(k, v []byte) error {
			var i apipb.Index
			var program cel.Program
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			if i.Connections {
				program, err = g.vm.Connection().Program(i.Expression)
				if err != nil {
					return err
				}
			} else if i.Docs {
				program, err = g.vm.Doc().Program(i.Expression)
				if err != nil {
					return err
				}
			}
			ind := &index{
				index:   &i,
				program: program,
			}
			g.indexes.Set(i.GetName(), ind, 0)
			return nil
		})
	})
}

func (g *Graph) rangeAuthorizers(fn func(a *authorizer) bool) {
	g.authorizers.Range(func(key, value interface{}) bool {
		return fn(value.(*authorizer))
	})
}

func (g *Graph) rangeTypeValidators(fn func(a *typeValidator) bool) {
	g.typeValidators.Range(func(key, value interface{}) bool {
		return fn(value.(*typeValidator))
	})
}

func (g *Graph) cacheAuthorizers() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbAuthorizers).ForEach(func(k, v []byte) error {
			if v == nil {
				return nil
			}
			var i apipb.Authorizer
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			program, err := g.vm.Auth().Program(i.Expression)
			if err != nil {
				return err
			}
			g.authorizers.Set(i.GetName(), &authorizer{
				authorizer: &i,
				program:    program,
			}, 0)
			return nil
		})
	})
}

func (g *Graph) cacheTypeValidators() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbTypeValidators).ForEach(func(k, v []byte) error {
			if v == nil {
				return nil
			}
			var i apipb.TypeValidator
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			var program cel.Program
			if i.GetConnections() {
				program, err = g.vm.Connection().Program(i.Expression)
				if err != nil {
					return err
				}
			}
			if i.GetDocs() {
				program, err = g.vm.Doc().Program(i.Expression)
				if err != nil {
					return err
				}
			}
			g.typeValidators.Set(i.GetName(), &typeValidator{
				validator: &i,
				program:   program,
			}, 0)
			return nil
		})
	})
}

func (g *Graph) setIndex(ctx context.Context, tx *bbolt.Tx, i *apipb.Index) (*apipb.Index, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	indexBucket := tx.Bucket(dbIndexes)
	if indexBucket == nil {
		return nil, errors.New("empty index bucket")
	}
	val := indexBucket.Get([]byte(i.GetName()))
	if val != nil && len(val) > 0 {
		var current = &apipb.Index{}
		err := proto.Unmarshal(val, current)
		if err != nil {
			return nil, err
		}
		current.Connections = i.Connections
		current.Docs = i.Docs
		current.Gtype = i.Gtype
		current.Expression = i.Expression
		bits, err := proto.Marshal(current)
		if err != nil {
			return nil, err
		}
		err = indexBucket.Put([]byte(i.GetName()), bits)
		if err != nil {
			return nil, err
		}
		return current, nil
	}
	bits, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := indexBucket.Put([]byte(i.GetName()), bits); err != nil {
		return nil, err
	}
	if i.Connections {
		tx.Bucket(dbIndexConnections).CreateBucketIfNotExists([]byte(i.GetName()))
	}
	if i.Docs {
		tx.Bucket(dbIndexDocs).CreateBucketIfNotExists([]byte(i.GetName()))
	}
	return i, nil
}

func (g *Graph) setAuthorizer(ctx context.Context, tx *bbolt.Tx, i *apipb.Authorizer) (*apipb.Authorizer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	authBucket := tx.Bucket(dbAuthorizers)
	val := authBucket.Get([]byte(i.GetName()))
	if val != nil && len(val) > 0 {
		var current = &apipb.Authorizer{}
		err := proto.Unmarshal(val, current)
		if err != nil {
			return nil, err
		}
		current.Expression = i.Expression
		bits, err := proto.Marshal(current)
		if err != nil {
			return nil, err
		}
		err = authBucket.Put([]byte(i.GetName()), bits)
		if err != nil {
			return nil, err
		}
		return current, nil
	}
	bits, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := authBucket.Put([]byte(i.GetName()), bits); err != nil {
		return nil, err
	}
	return i, nil
}

func (g *Graph) setTypedValidator(ctx context.Context, tx *bbolt.Tx, i *apipb.TypeValidator) (*apipb.TypeValidator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	validatorBucket := tx.Bucket(dbTypeValidators)
	val := validatorBucket.Get([]byte(i.GetName()))
	if val != nil && len(val) > 0 {
		var current = &apipb.TypeValidator{}
		err := proto.Unmarshal(val, current)
		if err != nil {
			return nil, err
		}
		current.Expression = i.Expression
		current.Docs = i.Docs
		current.Connections = i.Connections
		current.Gtype = i.Gtype
		bits, err := proto.Marshal(current)
		if err != nil {
			return nil, err
		}
		err = validatorBucket.Put([]byte(i.GetName()), bits)
		if err != nil {
			return nil, err
		}
		return current, nil
	}
	bits, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := validatorBucket.Put([]byte(i.GetName()), bits); err != nil {
		return nil, err
	}
	return i, nil
}

func (g *Graph) setIndexedDoc(ctx context.Context, tx *bbolt.Tx, index string, gid, doc []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bucket := tx.Bucket(dbIndexDocs).Bucket([]byte(index))
	if bucket == nil {
		return ErrNotFound
	}
	if err := bucket.Put(gid, doc); err != nil {
		return err
	}
	return nil
}

func (g *Graph) setIndexedConnection(ctx context.Context, tx *bbolt.Tx, index, gid, connection []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bucket := tx.Bucket(dbIndexConnections).Bucket([]byte(index))
	if bucket == nil {
		return ErrNotFound
	}
	if err := bucket.Put(gid, connection); err != nil {
		return err
	}
	return nil
}

func (g *Graph) delIndexedDoc(ctx context.Context, tx *bbolt.Tx, index, gid []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bucket := tx.Bucket(dbIndexDocs).Bucket(index)
	if bucket == nil {
		return ErrNotFound
	}
	if err := bucket.Delete(gid); err != nil {
		return err
	}
	return nil
}

func (g *Graph) delIndexedConnection(ctx context.Context, tx *bbolt.Tx, index, gid []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bucket := tx.Bucket(dbIndexConnections).Bucket(index)
	if bucket == nil {
		return ErrNotFound
	}
	if err := bucket.Delete(gid); err != nil {
		return err
	}
	return nil
}

func (g *Graph) setDoc(ctx context.Context, tx *bbolt.Tx, doc *apipb.Doc) (*apipb.Doc, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var validationErr error
	g.rangeTypeValidators(func(v *typeValidator) bool {
		if v.validator.GetDocs() && v.validator.GetGtype() == doc.GetRef().GetGtype() {
			res, err := g.vm.Doc().Eval(doc, v.program)
			if err != nil {
				validationErr = err
				return false
			}
			if !res {
				validationErr = errors.New(fmt.Sprintf("%s.%s document validation error! validator expression: %s", v.validator.GetGtype(), v.validator.GetName(), v.validator.GetExpression()))
				return false
			}
		}
		return true
	})
	if validationErr != nil {
		return nil, status.Error(codes.InvalidArgument, validationErr.Error())
	}
	bits, err := proto.Marshal(doc)
	if err != nil {
		return nil, err
	}
	docBucket := tx.Bucket(dbDocs)
	bucket := docBucket.Bucket([]byte(doc.GetRef().GetGtype()))
	if bucket == nil {
		bucket, err = docBucket.CreateBucketIfNotExists([]byte(doc.GetRef().GetGtype()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create bucket %s", doc.GetRef().GetGtype())
		}
	}
	if err := bucket.Put([]byte(doc.GetRef().GetGid()), bits); err != nil {
		return nil, err
	}
	g.rangeIndexes(func(i *index) bool {
		if i.index.Docs {
			result, err := g.vm.Doc().Eval(doc, i.program)
			if err != nil {
				if !strings.Contains(err.Error(), "no such key") {
					logger.Error("set index failure", zap.Error(err))
				}
			}
			if result {
				err = g.setIndexedDoc(ctx, tx, i.index.Name, []byte(doc.GetRef().GetGid()), bits)
				if err != nil {
					logger.Error("failed to save index", zap.Error(err))
					return true
				}
			}
		}
		return true
	})
	return doc, nil
}

func (g *Graph) setDocs(ctx context.Context, docs ...*apipb.Doc) (*apipb.Docs, error) {
	var nds = &apipb.Docs{}
	if err := g.db.Batch(func(tx *bbolt.Tx) error {
		for _, doc := range docs {
			n, err := g.setDoc(ctx, tx, doc)
			if err != nil {
				return err
			}
			nds.Docs = append(nds.Docs, n)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return nds, nil
}

func (g *Graph) setConnection(ctx context.Context, tx *bbolt.Tx, connection *apipb.Connection) (*apipb.Connection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	docBucket := tx.Bucket(dbDocs)
	{
		fromBucket := docBucket.Bucket([]byte(connection.GetFrom().GetGtype()))
		if fromBucket == nil {
			return nil, errors.Errorf("from doc %s does not exist", connection.GetFrom().String())
		}
		val := fromBucket.Get([]byte(connection.GetFrom().GetGid()))
		if val == nil {
			return nil, errors.Errorf("from doc %s does not exist", connection.GetFrom().String())
		}
	}
	{
		toBucket := docBucket.Bucket([]byte(connection.GetTo().GetGtype()))
		if toBucket == nil {
			return nil, errors.Errorf("to doc %s does not exist", connection.GetTo().String())
		}
		val := toBucket.Get([]byte(connection.GetTo().GetGid()))
		if val == nil {
			return nil, errors.Errorf("to doc %s does not exist", connection.GetTo().String())
		}
	}

	var validationErr error
	g.rangeTypeValidators(func(v *typeValidator) bool {
		if v.validator.GetConnections() && v.validator.GetGtype() == connection.GetRef().GetGtype() {
			res, err := g.vm.Connection().Eval(connection, v.program)
			if err != nil {
				validationErr = err
				return false
			}
			if !res {
				validationErr = errors.New(fmt.Sprintf("%s.%s connection validation error! validator expression: %s", v.validator.GetGtype(), v.validator.GetName(), v.validator.GetExpression()))
				return false
			}
		}
		return true
	})
	if validationErr != nil {
		return nil, status.Error(codes.InvalidArgument, validationErr.Error())
	}
	bits, err := proto.Marshal(connection)
	if err != nil {
		return nil, err
	}
	bucket := tx.Bucket(dbConnections)
	connectionBucket := bucket.Bucket([]byte(connection.GetRef().GetGtype()))
	if connectionBucket == nil {
		connectionBucket, err = bucket.CreateBucketIfNotExists([]byte(connection.GetRef().GetGtype()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create bucket %s", connection.GetRef().GetGtype())
		}
	}
	if err := connectionBucket.Put([]byte(connection.GetRef().GetGid()), bits); err != nil {
		return nil, err
	}
	refstr := refString(connection.GetRef())
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.connectionsFrom[connection.From.String()] == nil {
		g.connectionsFrom[connection.From.String()] = map[string]struct{}{}
	}
	if g.connectionsTo[connection.To.String()] == nil {
		g.connectionsTo[connection.To.String()] = map[string]struct{}{}
	}
	g.connectionsFrom[connection.GetFrom().String()][refstr] = struct{}{}
	g.connectionsTo[connection.GetTo().String()][refstr] = struct{}{}
	if !connection.Directed {
		g.connectionsTo[connection.GetFrom().String()][refstr] = struct{}{}
		g.connectionsFrom[connection.GetTo().String()][refstr] = struct{}{}
	}
	g.rangeIndexes(func(i *index) bool {
		if i.index.Connections {
			result, _ := g.vm.Connection().Eval(connection, i.program)
			if result {
				err = g.setIndexedConnection(ctx, tx, []byte(i.index.Name), []byte(connection.GetRef().GetGid()), bits)
				if err != nil {
					logger.Error("failed to save index", zap.Error(err))
					return true
				}
			}
		}
		return true
	})
	return connection, nil
}

func (g *Graph) setConnections(ctx context.Context, tx *bbolt.Tx, connections ...*apipb.Connection) (*apipb.Connections, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var edgs = &apipb.Connections{}
	for _, connection := range connections {
		e, err := g.setConnection(ctx, tx, connection)
		if err != nil {
			return nil, err
		}
		edgs.Connections = append(edgs.Connections, e)
	}
	return edgs, nil
}

func (g *Graph) getDoc(ctx context.Context, tx *bbolt.Tx, path *apipb.Ref) (*apipb.Doc, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if path == nil {
		return nil, ErrNotFound
	}
	var doc apipb.Doc
	docsBucket := tx.Bucket(dbDocs)
	bucket := docsBucket.Bucket([]byte(path.GetGtype()))
	if bucket == nil {
		return nil, ErrNotFound
	}
	bits := bucket.Get([]byte(path.Gid))
	if len(bits) == 0 {
		return nil, ErrNotFound
	}
	if err := proto.Unmarshal(bits, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (g *Graph) getConnection(ctx context.Context, tx *bbolt.Tx, path *apipb.Ref) (*apipb.Connection, error) {
	if path == nil {
		return nil, ErrNotFound
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var connection apipb.Connection
	bucket := tx.Bucket(dbConnections).Bucket([]byte(path.Gtype))
	if bucket == nil {
		return nil, ErrNotFound
	}
	bits := bucket.Get([]byte(path.Gid))
	if len(bits) == 0 {
		return nil, ErrNotFound
	}
	if err := proto.Unmarshal(bits, &connection); err != nil {
		return nil, err
	}
	return &connection, nil
}

func (g *Graph) rangeConnections(ctx context.Context, gType string, fn func(n *apipb.Connection) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := g.db.View(func(tx *bbolt.Tx) error {
		if gType == apipb.Any {
			types, err := g.ConnectionTypes(ctx)
			if err != nil {
				return err
			}
			for _, connectionType := range types {
				if err := g.rangeConnections(ctx, connectionType, fn); err != nil {
					return err
				}
			}
			return nil
		}
		bucket := tx.Bucket(dbConnections).Bucket([]byte(gType))
		if bucket == nil {
			return ErrNotFound
		}

		return bucket.ForEach(func(k, v []byte) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var connection apipb.Connection
			if err := proto.Unmarshal(v, &connection); err != nil {
				return err
			}
			if !fn(&connection) {
				return DONE
			}
			return nil
		})
	}); err != nil && err != DONE {
		return err
	}
	return nil
}

func (g *Graph) rangeDocs(ctx context.Context, gType string, fn func(n *apipb.Doc) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.db.View(func(tx *bbolt.Tx) error {
		if gType == apipb.Any {
			types, err := g.DocTypes(ctx)
			if err != nil {
				return err
			}
			for _, docType := range types {
				if err := g.rangeDocs(ctx, docType, fn); err != nil {
					return err
				}
			}
			return nil
		}
		bucket := tx.Bucket(dbDocs).Bucket([]byte(gType))
		if bucket == nil {
			return ErrNotFound
		}

		return bucket.ForEach(func(k, v []byte) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var doc apipb.Doc
			if err := proto.Unmarshal(v, &doc); err != nil {
				return err
			}
			if !fn(&doc) {
				return DONE
			}
			return nil
		})
	}); err != nil && err != DONE {
		return err
	}
	return nil
}

func (g *Graph) createIdentity(ctx context.Context, constructor *apipb.DocConstructor) (*apipb.Doc, error) {

	var (
		err     error
		newDock *apipb.Doc
	)

	if err := g.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket(dbDocs)
		bucket := docBucket.Bucket([]byte(constructor.GetRef().GetGtype()))
		if bucket == nil {
			bucket, err = docBucket.CreateBucket([]byte(constructor.GetRef().GetGtype()))
			if err != nil {
				return errors.Wrapf(err, "%s", constructor.GetRef().GetGtype())
			}
		}
		path := &apipb.Ref{Gid: constructor.GetRef().GetGid(), Gtype: constructor.GetRef().GetGtype()}
		newDock = &apipb.Doc{
			Ref:        path,
			Attributes: constructor.GetAttributes(),
		}
		newDock, err = g.setDoc(ctx, tx, newDock)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return newDock, nil
}

func (g *Graph) delDoc(ctx context.Context, tx *bbolt.Tx, path *apipb.Ref) error {
	doc, err := g.getDoc(ctx, tx, path)
	if err != nil {
		return err
	}
	bucket := tx.Bucket(dbDocs).Bucket([]byte(doc.GetRef().GetGtype()))
	g.rangeFrom(ctx, tx, path, func(e *apipb.Connection) bool {
		g.delConnection(ctx, tx, e.GetRef())
		return true
	})
	g.rangeTo(ctx, tx, path, func(e *apipb.Connection) bool {
		g.delConnection(ctx, tx, e.GetRef())
		return true
	})
	g.rangeIndexes(func(index *index) bool {
		if index.index.Docs && index.index.GetGtype() == path.GetGtype() {
			g.delIndexedDoc(ctx, tx, []byte(index.index.Name), []byte(path.GetGid()))
		}
		return true
	})
	return bucket.Delete([]byte(path.GetGid()))
}

func (g *Graph) delConnection(ctx context.Context, tx *bbolt.Tx, path *apipb.Ref) error {
	connection, err := g.getConnection(ctx, tx, path)
	if err != nil {
		return err
	}
	g.mu.Lock()
	if g.connectionsFrom != nil {
		delete(g.connectionsFrom[connection.GetFrom().String()], refString(path))
	}
	if g.connectionsTo != nil {
		delete(g.connectionsTo[connection.GetTo().String()], refString(path))
	}
	g.mu.Unlock()
	g.rangeIndexes(func(index *index) bool {
		if index.index.Connections && index.index.GetGtype() == path.GetGtype() {
			g.delIndexedConnection(ctx, tx, []byte(index.index.Name), []byte(path.GetGid()))
		}
		return true
	})
	return tx.Bucket(dbConnections).Bucket([]byte(connection.GetRef().GetGtype())).Delete([]byte(connection.GetRef().GetGid()))
}

func (n *Graph) filterDoc(ctx context.Context, docType string, filter func(doc *apipb.Doc) bool) (*apipb.Docs, error) {
	var filtered []*apipb.Doc
	if err := n.rangeDocs(ctx, docType, func(doc *apipb.Doc) bool {
		if filter(doc) {
			filtered = append(filtered, doc)
		}
		return true
	}); err != nil {
		return nil, err
	}
	toreturn := &apipb.Docs{
		Docs: filtered,
	}
	return toreturn, nil
}

func (g *Graph) rangeTo(ctx context.Context, tx *bbolt.Tx, docRef *apipb.Ref, fn func(e *apipb.Connection) bool) error {
	g.mu.RLock()
	pathMap := g.connectionsTo[docRef.String()]
	g.mu.RUnlock()
	var paths []string
	for path, _ := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		var ref apipb.Ref
		proto.Unmarshal([]byte(path), &ref)
		bucket := tx.Bucket(dbConnections).Bucket([]byte(ref.GetGtype()))
		if bucket == nil {
			return ErrNotFound
		}
		var connection apipb.Connection
		bits := bucket.Get([]byte(ref.GetGid()))
		if err := proto.Unmarshal(bits, &connection); err != nil {
			return err
		}
		if !fn(&connection) {
			return nil
		}
	}
	return nil
}

func (g *Graph) rangeFrom(ctx context.Context, tx *bbolt.Tx, docRef *apipb.Ref, fn func(e *apipb.Connection) bool) error {
	g.mu.RLock()
	pathMap := g.connectionsFrom[docRef.String()]
	g.mu.RUnlock()
	var paths []string
	for path, _ := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, val := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := fromRefString(val)
		bucket := tx.Bucket(dbConnections).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var connection apipb.Connection
		bits := bucket.Get([]byte(path.GetGid()))
		if err := proto.Unmarshal(bits, &connection); err != nil {
			return err
		}
		if !fn(&connection) {
			return nil
		}
	}
	return nil
}

func (g *Graph) rangeSeekConnections(ctx context.Context, gType string, seek string, index string, reverse bool, fn func(e *apipb.Connection) bool) (string, error) {
	if ctx.Err() != nil {
		return seek, ctx.Err()
	}
	var lastKey []byte
	if err := g.db.View(func(tx *bbolt.Tx) error {
		var c *bbolt.Cursor
		if index != "" {
			bucket := tx.Bucket(dbIndexConnections).Bucket([]byte(index))
			if bucket == nil {
				return ErrNotFound
			}
			c = bucket.Cursor()
		} else {
			bucket := tx.Bucket(dbConnections).Bucket([]byte(gType))
			if bucket == nil {
				return ErrNotFound
			}
			c = bucket.Cursor()
		}
		var iter = c.Next
		if reverse {
			iter = c.Prev
		}
		if seek != "" {
			for k, v := c.Seek([]byte(seek)); k != nil; k, v = iter() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				var connection apipb.Connection
				if err := proto.Unmarshal(v, &connection); err != nil {
					return err
				}
				lastKey = k
				if !fn(&connection) {
					return DONE
				}
			}
		} else {
			for k, v := c.First(); k != nil; k, v = iter() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				var connection apipb.Connection
				if err := proto.Unmarshal(v, &connection); err != nil {
					return err
				}
				lastKey = k
				if !fn(&connection) {
					return DONE
				}
			}
		}
		return nil
	}); err != nil && err != DONE {
		return string(lastKey), err
	}
	return string(lastKey), nil
}

func (g *Graph) rangeSeekDocs(ctx context.Context, gType string, seek string, index string, reverse bool, fn func(e *apipb.Doc) bool) (string, error) {
	if ctx.Err() != nil {
		return seek, ctx.Err()
	}
	var lastKey []byte
	if err := g.db.View(func(tx *bbolt.Tx) error {
		var c *bbolt.Cursor
		if index != "" {
			bucket := tx.Bucket(dbIndexDocs).Bucket([]byte(index))
			if bucket == nil {
				return ErrNotFound
			}
			c = bucket.Cursor()
		} else {
			bucket := tx.Bucket(dbDocs).Bucket([]byte(gType))
			if bucket == nil {
				return ErrNotFound
			}
			c = bucket.Cursor()
		}
		var iter = c.Next
		if reverse {
			iter = c.Prev
		}
		if seek != "" {
			for k, v := c.Seek([]byte(seek)); k != nil; k, v = iter() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				var doc apipb.Doc
				if err := proto.Unmarshal(v, &doc); err != nil {
					return err
				}
				lastKey = k
				if !fn(&doc) {
					return DONE
				}
			}
		} else {
			for k, v := c.First(); k != nil; k, v = iter() {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				var doc apipb.Doc
				if err := proto.Unmarshal(v, &doc); err != nil {
					return err
				}
				lastKey = k
				if !fn(&doc) {
					return DONE
				}
			}
		}
		return nil
	}); err != nil && err != DONE {
		return string(lastKey), err
	}
	return string(lastKey), nil
}

func (g *Graph) hasConnectionFrom(doc, connection *apipb.Ref) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	values := g.connectionsFrom[doc.String()]
	if values == nil {
		return false
	}
	for path, _ := range values {
		val := fromRefString(path)
		if val.Gid == connection.Gid && val.GetGtype() == connection.GetGtype() {
			return true
		}
	}
	return false
}

func (g *Graph) hasConnectionTo(doc, connection *apipb.Ref) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	values := g.connectionsTo[doc.String()]
	if values == nil {
		return false
	}
	for path, _ := range values {
		val := fromRefString(path)
		if val.Gid == connection.Gid && val.GetGtype() == connection.GetGtype() {
			return true
		}
	}
	return false
}
