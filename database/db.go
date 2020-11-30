package database

import (
	"context"
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/autom8ter/graphik/helpers"
	"github.com/autom8ter/graphik/logger"
	"github.com/google/cel-go/cel"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
)

func (g *Graph) updateMeta(ctx context.Context, meta *apipb.Metadata) {
	identity := g.getIdentity(ctx)
	if meta == nil {
		meta = &apipb.Metadata{}
	}
	if meta.GetCreatedAt() == nil {
		meta.CreatedAt = timestamppb.Now()
	}
	if meta.GetCreatedBy() == nil {
		meta.CreatedBy = identity.GetPath()
	}
	if identity != nil {
		meta.UpdatedBy = identity.GetPath()
	}

	meta.UpdatedAt = timestamppb.Now()

	meta.Version += 1
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

func (g *Graph) setIndex(ctx context.Context, tx *bbolt.Tx, i *apipb.Index) (*apipb.Index, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	indexBucket := tx.Bucket(dbIndexes)
	val := indexBucket.Get([]byte(i.GetName()))
	if val != nil {
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
		return current, g.cacheIndexes()
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
	return i, g.cacheIndexes()
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
	if doc.GetPath() == nil {
		doc.Path = &apipb.Path{}
	}
	g.updateMeta(ctx, doc.Metadata)
	bits, err := proto.Marshal(doc)
	if err != nil {
		return nil, err
	}
	docBucket := tx.Bucket(dbDocs)
	bucket := docBucket.Bucket([]byte(doc.GetPath().GetGtype()))
	if bucket == nil {
		bucket, err = docBucket.CreateBucketIfNotExists([]byte(doc.GetPath().GetGtype()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create bucket %s", doc.GetPath().GetGtype())
		}
	}
	if err := bucket.Put(helpers.Uint64ToBytes(uint64(doc.GetPath().GetGid())), bits); err != nil {
		return nil, err
	}
	g.rangeIndexes(func(i *index) bool {
		if i.index.Docs {
			program, err := g.vm.Doc().Program(i.index.Expression)
			if err != nil {
				logger.Error("failed to create index", zap.Error(err))
				return true
			}
			result, err := g.vm.Doc().Eval(doc, program)
			if err != nil {
				logger.Error("failed to evaluate index", zap.Error(err))
				return true
			}
			if result {
				err = g.setIndexedDoc(ctx, tx, i.index.Name, helpers.Uint64ToBytes(uint64(doc.GetPath().GetGid())), bits)
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
	if err := g.db.Update(func(tx *bbolt.Tx) error {
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
		fromBucket := docBucket.Bucket([]byte(connection.From.Gtype))
		if fromBucket == nil {
			return nil, errors.Errorf("from doc %s does not exist", connection.From.String())
		}
		val := fromBucket.Get(helpers.Uint64ToBytes(uint64(connection.From.Gid)))
		if val == nil {
			return nil, errors.Errorf("from doc %s does not exist", connection.From.String())
		}
	}
	{
		toBucket := docBucket.Bucket([]byte(connection.To.Gtype))
		if toBucket == nil {
			return nil, errors.Errorf("to doc %s does not exist", connection.To.String())
		}
		val := toBucket.Get(helpers.Uint64ToBytes(uint64(connection.To.Gid)))
		if val == nil {
			return nil, errors.Errorf("to doc %s does not exist", connection.To.String())
		}
	}
	if connection.GetPath() == nil {
		connection.Path = &apipb.Path{}
	}
	g.updateMeta(ctx, connection.Metadata)
	bits, err := proto.Marshal(connection)
	if err != nil {
		return nil, err
	}
	connectionBucket := tx.Bucket(dbConnections)
	connectionBucket = connectionBucket.Bucket([]byte(connection.GetPath().GetGtype()))
	if connectionBucket == nil {
		connectionBucket, err = connectionBucket.CreateBucketIfNotExists([]byte(connection.GetPath().GetGtype()))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create bucket %s", connection.GetPath().GetGtype())
		}
	}
	if err := connectionBucket.Put(helpers.Uint64ToBytes(uint64(connection.GetPath().GetGid())), bits); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.connectionsFrom[connection.GetFrom().String()] = append(g.connectionsFrom[connection.GetFrom().String()], connection.GetPath())
	g.connectionsTo[connection.GetTo().String()] = append(g.connectionsTo[connection.GetTo().String()], connection.GetPath())
	if !connection.Directed {
		g.connectionsTo[connection.GetFrom().String()] = append(g.connectionsTo[connection.GetFrom().String()], connection.GetPath())
		g.connectionsFrom[connection.GetTo().String()] = append(g.connectionsFrom[connection.GetTo().String()], connection.GetPath())
	}
	sortPaths(g.connectionsFrom[connection.GetFrom().String()])
	sortPaths(g.connectionsTo[connection.GetTo().String()])
	g.rangeIndexes(func(i *index) bool {
		if i.index.Connections {
			program, err := g.vm.Connection().Program(i.index.Expression)
			if err != nil {
				logger.Error("failed to create index", zap.Error(err))
				return true
			}
			result, err := g.vm.Connection().Eval(connection, program)
			if err != nil {
				logger.Error("failed to evaluate index", zap.Error(err))
				return true
			}
			if result {
				err = g.setIndexedConnection(ctx, tx, []byte(i.index.Name), helpers.Uint64ToBytes(uint64(connection.GetPath().GetGid())), bits)
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

func (g *Graph) setConnections(ctx context.Context, connections ...*apipb.Connection) (*apipb.Connections, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var edgs = &apipb.Connections{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, connection := range connections {
			e, err := g.setConnection(ctx, tx, connection)
			if err != nil {
				return err
			}
			edgs.Connections = append(edgs.Connections, e)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return edgs, nil
}

func (g *Graph) getDoc(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) (*apipb.Doc, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var doc apipb.Doc
	bucket := tx.Bucket(dbDocs).Bucket([]byte(path.Gtype))
	if bucket == nil {
		return nil, ErrNotFound
	}
	bits := bucket.Get(helpers.Uint64ToBytes(uint64(path.Gid)))
	if len(bits) == 0 {
		return nil, ErrNotFound
	}
	if err := proto.Unmarshal(bits, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (g *Graph) getConnection(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) (*apipb.Connection, error) {
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
	bits := bucket.Get(helpers.Uint64ToBytes(uint64(path.Gid)))
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
		now     = timestamppb.Now()
		newDock *apipb.Doc
	)

	if err := g.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket(dbDocs)
		bucket := docBucket.Bucket([]byte(constructor.GetGtype()))
		if bucket == nil {
			bucket, err = docBucket.CreateBucket([]byte(constructor.GetGtype()))
			if err != nil {
				return errors.Wrapf(err, "%s", constructor.GetGtype())
			}
		}
		seq, err := bucket.NextSequence()
		if err != nil {
			return errors.Wrap(err, "failed to get next sequence")
		}
		newDock = &apipb.Doc{
			Path: &apipb.Path{
				Gtype: constructor.GetGtype(),
				Gid:   int64(seq),
			},
			Attributes: constructor.GetAttributes(),
			Metadata: &apipb.Metadata{
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: &apipb.Path{
					Gtype: constructor.GetGtype(),
					Gid:   int64(seq),
				},
				UpdatedBy: &apipb.Path{
					Gtype: constructor.GetGtype(),
					Gid:   int64(seq),
				},
			},
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

func (g *Graph) delDoc(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) error {
	doc, err := g.getDoc(ctx, tx, path)
	if err != nil {
		return err
	}
	bucket := tx.Bucket(dbDocs).Bucket([]byte(doc.GetPath().GetGtype()))
	idBuf := helpers.Uint64ToBytes(uint64(doc.GetPath().GetGid()))
	g.rangeFrom(ctx, tx, path, func(e *apipb.Connection) bool {
		g.delConnection(ctx, tx, e.GetPath())
		return true
	})
	g.rangeTo(ctx, tx, path, func(e *apipb.Connection) bool {
		g.delConnection(ctx, tx, e.GetPath())
		return true
	})
	g.rangeIndexes(func(index *index) bool {
		if index.index.Docs && index.index.GetGtype() == path.GetGtype() {
			g.delIndexedDoc(ctx, tx, []byte(index.index.Name), idBuf)
		}
		return true
	})
	return bucket.Delete(idBuf)
}

func (g *Graph) delConnection(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) error {
	connection, err := g.getConnection(ctx, tx, path)
	if err != nil {
		return err
	}
	g.mu.Lock()
	fromPaths := removeConnection(path, g.connectionsFrom[connection.GetFrom().String()])
	g.connectionsFrom[connection.GetFrom().String()] = fromPaths
	toPaths := removeConnection(path, g.connectionsTo[connection.GetTo().String()])
	g.connectionsTo[connection.GetTo().String()] = toPaths
	g.mu.Unlock()
	g.rangeIndexes(func(index *index) bool {
		if index.index.Connections && index.index.GetGtype() == path.GetGtype() {
			g.delIndexedConnection(ctx, tx, []byte(index.index.Name), helpers.Uint64ToBytes(uint64(path.GetGid())))
		}
		return true
	})
	return tx.Bucket(dbConnections).Bucket([]byte(connection.GetPath().GetGtype())).Delete(helpers.Uint64ToBytes(uint64(connection.GetPath().GetGid())))
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

func (g *Graph) rangeTo(ctx context.Context, tx *bbolt.Tx, docPath *apipb.Path, fn func(e *apipb.Connection) bool) error {
	g.mu.RLock()
	paths := g.connectionsTo[docPath.String()]
	g.mu.RUnlock()
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := tx.Bucket(dbConnections).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var connection apipb.Connection
		bits := bucket.Get(helpers.Uint64ToBytes(uint64(path.Gid)))
		if err := proto.Unmarshal(bits, &connection); err != nil {
			return err
		}
		if !fn(&connection) {
			return nil
		}
	}
	return nil
}

func (g *Graph) rangeFrom(ctx context.Context, tx *bbolt.Tx, docPath *apipb.Path, fn func(e *apipb.Connection) bool) error {
	g.mu.RLock()
	paths := g.connectionsFrom[docPath.String()]
	g.mu.RUnlock()
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := tx.Bucket(dbConnections).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		var connection apipb.Connection
		bits := bucket.Get(helpers.Uint64ToBytes(uint64(path.Gid)))
		if err := proto.Unmarshal(bits, &connection); err != nil {
			return err
		}
		if !fn(&connection) {
			return nil
		}
	}
	return nil
}

func (g *Graph) rangeSeekConnections(ctx context.Context, gType string, seek int64, index string, reverse bool, fn func(e *apipb.Connection) bool) (int64, error) {
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
		for k, v := c.Seek(helpers.Uint64ToBytes(uint64(seek))); k != nil; k, v = iter() {
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
		return nil
	}); err != nil && err != DONE {
		return 0, err
	}
	return int64(helpers.BytesToUint64(lastKey)), nil
}

func (g *Graph) rangeSeekDocs(ctx context.Context, gType string, seek int64, index string, reverse bool, fn func(e *apipb.Doc) bool) (int64, error) {
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
		for k, v := c.Seek(helpers.Uint64ToBytes(uint64(seek))); k != nil; k, v = iter() {
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
		return nil
	}); err != nil && err != DONE {
		return int64(helpers.BytesToUint64(lastKey)), err
	}
	return int64(helpers.BytesToUint64(lastKey)), nil
}

func removeConnection(path *apipb.Path, paths []*apipb.Path) []*apipb.Path {
	var newPaths []*apipb.Path
	for _, p := range paths {
		if path.Gid == p.Gid && path.Gtype == p.Gtype {
			newPaths = append(newPaths, p)
		}
	}
	sortPaths(newPaths)
	return newPaths
}

func sortPaths(paths []*apipb.Path) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].String() < paths[j].String()
	})
}

type index struct {
	index   *apipb.Index
	program cel.Program
}
