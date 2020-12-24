package database

import (
	"context"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"github.com/graphikDB/trigger"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
)

type index struct {
	index    *apipb.Index
	decision *trigger.Decision
}

type authorizer struct {
	authorizer *apipb.Authorizer
	decision   *trigger.Decision
}

type typeValidator struct {
	validator *apipb.Constraint
	decision  *trigger.Decision
}

type triggerCache struct {
	trigger     *apipb.Trigger
	evalTrigger *trigger.Trigger
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
	return g.rangeConnections(context.WithValue(context.Background(), bypassAuthorizersCtxKey, true), apipb.Any, func(e *apipb.Connection) bool {
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
			if k == nil || v == nil {
				return nil
			}
			var i apipb.Index
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			decision, err := trigger.NewDecision(i.Expression)
			if err != nil {
				return err
			}
			ind := &index{
				index:    &i,
				decision: decision,
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

func (g *Graph) rangeTriggers(fn func(a *triggerCache) bool) {
	g.triggers.Range(func(key, value interface{}) bool {
		return fn(value.(*triggerCache))
	})
}

func (g *Graph) rangeConstraints(fn func(a *typeValidator) bool) {
	g.typeValidators.Range(func(key, value interface{}) bool {
		return fn(value.(*typeValidator))
	})
}

func (g *Graph) cacheAuthorizers() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbAuthorizers).ForEach(func(k, v []byte) error {
			if k == nil || v == nil {
				return nil
			}
			var i apipb.Authorizer
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			if i.GetExpression() == "" {
				return nil
			}
			decision, err := trigger.NewDecision(i.GetExpression())
			if err != nil {
				return errors.Wrapf(err, "failed to cache auth expression: %s", i.GetName())
			}
			g.authorizers.Set(i.GetName(), &authorizer{
				authorizer: &i,
				decision:   decision,
			}, 0)
			return nil
		})
	})
}

func (g *Graph) cacheTriggers() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbTriggers).ForEach(func(k, v []byte) error {
			if k == nil || v == nil {
				return nil
			}
			var i apipb.Trigger
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			if i.GetExpression() == "" {
				return nil
			}
			decision, err := trigger.NewDecision(i.GetExpression())
			if err != nil {
				return errors.Wrapf(err, "failed to cache trigger decision: %s", i.GetName())
			}
			trig, err := trigger.NewTrigger(decision, i.GetTrigger())
			if err != nil {
				return errors.Wrapf(err, "failed to cache trigger expression: %s", i.GetName())
			}
			g.triggers.Set(i.GetName(), &triggerCache{
				trigger:     &i,
				evalTrigger: trig,
			}, 0)
			return nil
		})
	})
}

func (g *Graph) cacheConstraints() error {
	return g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbConstraints).ForEach(func(k, v []byte) error {
			if k == nil || v == nil {
				return nil
			}
			var i apipb.Constraint
			var err error
			if err := proto.Unmarshal(v, &i); err != nil {
				return err
			}
			decision, err := trigger.NewDecision(i.Expression)
			if err != nil {
				return err
			}
			g.typeValidators.Set(i.GetName(), &typeValidator{
				validator: &i,
				decision:  decision,
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
		current.TargetResponses = i.TargetResponses
		current.TargetRequests = i.TargetRequests
		current.Method = i.Method
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

func (g *Graph) setConstraint(ctx context.Context, tx *bbolt.Tx, i *apipb.Constraint) (*apipb.Constraint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	validatorBucket := tx.Bucket(dbConstraints)
	val := validatorBucket.Get([]byte(i.GetName()))
	if val != nil && len(val) > 0 {
		var current = &apipb.Constraint{}
		err := proto.Unmarshal(val, current)
		if err != nil {
			return nil, err
		}
		current.Expression = i.Expression
		current.TargetDocs = i.TargetDocs
		current.TargetConnections = i.TargetConnections
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

func (g *Graph) setTrigger(ctx context.Context, tx *bbolt.Tx, i *apipb.Trigger) (*apipb.Trigger, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	triggerBucket := tx.Bucket(dbTriggers)
	val := triggerBucket.Get([]byte(i.GetName()))
	if val != nil && len(val) > 0 {
		var current = &apipb.Trigger{}
		err := proto.Unmarshal(val, current)
		if err != nil {
			return nil, err
		}
		current.Expression = i.Expression
		current.TargetDocs = i.TargetDocs
		current.TargetConnections = i.TargetConnections
		current.Gtype = i.Gtype
		current.Trigger = i.Trigger
		bits, err := proto.Marshal(current)
		if err != nil {
			return nil, err
		}
		err = triggerBucket.Put([]byte(i.GetName()), bits)
		if err != nil {
			return nil, err
		}
		return current, nil
	}
	bits, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	if err := triggerBucket.Put([]byte(i.GetName()), bits); err != nil {
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
	docMap := doc.AsMap()
	var validationErr error
	g.rangeConstraints(func(v *typeValidator) bool {
		if v.validator.GetTargetDocs() && (v.validator.GetGtype() == apipb.Any || v.validator.GetGtype() == doc.GetRef().GetGtype()) {
			err := v.decision.Eval(docMap)
			if err != nil {
				validationErr = errors.Wrapf(err, "%s.%s document validation error! validator expression: %s", v.validator.GetGtype(), v.validator.GetName(), v.validator.GetExpression())
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
		if i.index.GetDocs() && i.index.GetGtype() == doc.GetRef().GetGtype() {
			if err := i.decision.Eval(docMap); err == nil {
				err = g.setIndexedDoc(ctx, tx, i.index.Name, []byte(doc.GetRef().GetGid()), bits)
				if err != nil {
					g.logger.Error("failed to save index", zap.Error(err))
					return true
				}
			}
		}
		return true
	})
	if err := g.machine.PubSub().Publish(changeChannel, &apipb.Message{
		Channel:   changeChannel,
		Data:      apipb.NewStruct(doc.AsMap()),
		User:      g.getIdentity(ctx).GetRef(),
		Timestamp: timestamppb.Now(),
		Method:    g.getMethod(ctx),
	}); err != nil {
		return nil, err
	}
	return doc, nil
}

func (g *Graph) setDocs(ctx context.Context, tx *bbolt.Tx, docs ...*apipb.Doc) (*apipb.Docs, error) {
	var nds = &apipb.Docs{}
	for _, doc := range docs {
		n, err := g.setDoc(ctx, tx, doc)
		if err != nil {
			return nil, err
		}
		nds.Docs = append(nds.Docs, n)
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
		if val == nil || len(val) == 0 {
			return nil, errors.Errorf("from doc %s does not exist", connection.GetFrom().String())
		}
	}
	{
		toBucket := docBucket.Bucket([]byte(connection.GetTo().GetGtype()))
		if toBucket == nil {
			return nil, errors.Errorf("to doc %s does not exist", connection.GetTo().String())
		}
		val := toBucket.Get([]byte(connection.GetTo().GetGid()))
		if val == nil || len(val) == 0 {
			return nil, errors.Errorf("to doc %s does not exist", connection.GetTo().String())
		}
	}
	connMap := connection.AsMap()
	var validationErr error
	g.rangeConstraints(func(v *typeValidator) bool {
		if v.validator.GetTargetConnections() && (v.validator.GetGtype() == apipb.Any || v.validator.GetGtype() == connection.GetRef().GetGtype()) {
			err := v.decision.Eval(connMap)
			if err != nil {
				validationErr = errors.Wrapf(err, "%s.%s connection validation error! validator expression: %s", v.validator.GetGtype(), v.validator.GetName(), v.validator.GetExpression())
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
	if g.connectionsFrom[connection.GetFrom().String()] == nil {
		g.connectionsFrom[connection.GetFrom().String()] = map[string]struct{}{}
	}
	if g.connectionsTo[connection.GetTo().String()] == nil {
		g.connectionsTo[connection.GetTo().String()] = map[string]struct{}{}
	}
	g.connectionsFrom[connection.GetFrom().String()][refstr] = struct{}{}
	g.connectionsTo[connection.GetTo().String()][refstr] = struct{}{}
	if !connection.Directed {
		g.connectionsTo[connection.GetFrom().String()][refstr] = struct{}{}
		g.connectionsFrom[connection.GetTo().String()][refstr] = struct{}{}
	}
	g.rangeIndexes(func(i *index) bool {
		if i.index.Connections && i.index.GetGtype() == connection.GetRef().GetGtype() {
			if err := i.decision.Eval(connMap); err == nil {
				err = g.setIndexedConnection(ctx, tx, []byte(i.index.Name), []byte(connection.GetRef().GetGid()), bits)
				if err != nil {
					g.logger.Error("failed to save index", zap.Error(err))
					return true
				}
			}
		}
		return true
	})
	if err := g.machine.PubSub().Publish(changeChannel, &apipb.Message{
		Channel:   changeChannel,
		Data:      apipb.NewStruct(connection.AsMap()),
		User:      g.getIdentity(ctx).GetRef(),
		Timestamp: timestamppb.Now(),
		Method:    g.getMethod(ctx),
	}); err != nil {
		return nil, err
	}
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
		g.rangeTriggers(func(a *triggerCache) bool {
			if a.trigger.GetTargetDocs() && newDock.GetRef().GetGtype() == a.trigger.GetGtype() {
				data, err := a.evalTrigger.Trigger(newDock.AsMap())
				if err == nil && len(data) > 0 {
					for k, v := range data {
						val, _ := structpb.NewValue(v)
						newDock.GetAttributes().GetFields()[k] = val
					}
				}
			}
			return true
		})
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
		if index.index.Docs && index.index.GetGtype() == doc.GetRef().GetGtype() {
			g.delIndexedDoc(ctx, tx, []byte(index.index.Name), []byte(path.GetGid()))
		}
		return true
	})
	if err := g.machine.PubSub().Publish(changeChannel, &apipb.Message{
		Channel:   changeChannel,
		Data:      apipb.NewStruct(path.AsMap()),
		User:      g.getIdentity(ctx).GetRef(),
		Timestamp: timestamppb.Now(),
		Method:    g.getMethod(ctx),
	}); err != nil {
		return err
	}
	if err := bucket.Delete([]byte(path.GetGid())); err != nil {
		return err
	}
	return nil
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
	if err := tx.Bucket(dbConnections).Bucket([]byte(connection.GetRef().GetGtype())).Delete([]byte(connection.GetRef().GetGid())); err != nil {
		return err
	}
	if err := g.machine.PubSub().Publish(changeChannel, &apipb.Message{
		Channel:   changeChannel,
		Data:      apipb.NewStruct(path.AsMap()),
		User:      g.getIdentity(ctx).GetRef(),
		Timestamp: timestamppb.Now(),
		Method:    g.getMethod(ctx),
	}); err != nil {
		return err
	}
	return nil
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
