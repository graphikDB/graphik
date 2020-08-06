package boltdb

import (
	"errors"
	"fmt"
	"github.com/autom8ter/graphik"
	"go.etcd.io/bbolt"
	"os"
	"time"
)

type Graph struct {
	path            string
	db              *bbolt.DB
	onNodeChange    []graphik.NodeTriggerFunc
	onEdgeChange    []graphik.EdgeTriggerFunc
	nodeConstraints []graphik.NodeConstraintFunc
	edgeConstraints []graphik.EdgeConstraintFunc
}

func (g *Graph) EdgeTriggers(changeHandlers ...graphik.EdgeTriggerFunc) {
	g.onEdgeChange = append(g.onEdgeChange, changeHandlers...)
}

func (g *Graph) NodeTriggers(changeHandlers ...graphik.NodeTriggerFunc) {
	g.onNodeChange = append(g.onNodeChange, changeHandlers...)
}

func (g *Graph) EdgeConstraints(constraints ...graphik.EdgeConstraintFunc) {
	g.edgeConstraints = append(g.edgeConstraints, constraints...)
}

func (g *Graph) NodeConstraints(constraints ...graphik.NodeConstraintFunc) {
	g.nodeConstraints = append(g.nodeConstraints, constraints...)
}

const DefaultPath = "/tmp/graphik"

// Open returns a Graph.
func Open(path string) graphik.GraphOpenerFunc {
	return func() (graphik.Graph, error) {
		if path == "" {
			path = DefaultPath
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0777)
		}
		db, err := bbolt.Open(fmt.Sprintf("%s/graph.boltdb", path), 0777, nil)

		if err != nil {
			return nil, err
		}
		return &Graph{
			path: path,
			db:   db,
		}, nil
	}
}

func (g *Graph) implements() graphik.Graph {
	return g
}

func (g *Graph) Close() error {
	return g.db.Close()
}

func (g *Graph) AddNode(n graphik.Node) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists([]byte(n.Type()))
		bucket, _ = bucket.CreateBucketIfNotExists([]byte(n.Key()))
		n.SetAttribute("last_update", time.Now().Unix())
		for _, fn := range g.nodeConstraints {
			if err := fn(g, n, false); err != nil {
				return err
			}
		}
		bits, err := n.Marshal()
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte("attributes"), bits); err != nil {
			return err
		}
		for _, fn := range g.onNodeChange {
			if err := fn(g, n, false); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *Graph) QueryNodes(query graphik.NodeQuery) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		if query == nil {
			query = graphik.NewNodeQuery()
		}
		count := 0
		fromTypes, err := fromTypeBuckets(tx, query.Type())
		if err != nil {
			return err
		}
		for _, fromType := range fromTypes.buckets {
			fromKeys, err := fromKeyBuckets(fromType, query.Key())
			if err != nil {
				return err
			}
			for _, fromKey := range fromKeys.buckets {
				if query.Limit() > 0 && count >= query.Limit() {
					return nil
				}
				var n = graphik.NewNode(graphik.NewPath(fromTypes.key, fromKeys.key), nil)
				res := fromKey.Get([]byte("attributes"))
				if len(res) > 0 {
					if err := n.Unmarshal(res); err != nil {
						return err
					}
					if query.Where() != nil {
						if query.Where()(g, n) {
							if err := query.Handler()(g, n); err != nil {
								return err
							}
							count++
						}
					} else {
						if err := query.Handler()(g, n); err != nil {
							return err
						}
						count++
					}
				}
			}
		}
		return nil
	})
}

func (g *Graph) GetNode(path graphik.Path) (graphik.Node, error) {
	var node = graphik.NewNode(path, nil)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(path.Type()))
		if bucket == nil {
			return errors.New("not found")
		}
		bucket = bucket.Bucket([]byte(path.Key()))
		if bucket == nil {
			return errors.New("not found")
		}
		res := bucket.Get([]byte("attributes"))
		if err := node.Unmarshal(res); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (g *Graph) DelNode(path graphik.Path) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		n, err := g.GetNode(path)
		if err != nil {
			return err
		}
		for _, fn := range g.nodeConstraints {
			if err := fn(g, n, true); err != nil {
				return err
			}
		}
		bucket := tx.Bucket([]byte(path.Type()))
		if err := bucket.Delete([]byte(path.Key())); err != nil {
			return err
		}
		for _, fn := range g.onNodeChange {
			if err := fn(g, n, true); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *Graph) AddEdge(e graphik.Edge) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, fn := range g.edgeConstraints {
			if err := fn(g, e, false); err != nil {
				return err
			}
		}
		bits, err := e.Marshal()
		if err != nil {
			return err
		}
		bucket, _ := tx.CreateBucketIfNotExists([]byte(e.From().Type()))
		bucket, _ = bucket.CreateBucketIfNotExists([]byte(e.From().Key()))
		bucket, _ = bucket.CreateBucketIfNotExists([]byte(e.Relationship()))
		bucket, _ = bucket.CreateBucketIfNotExists([]byte(e.To().Type()))
		bucket, _ = bucket.CreateBucketIfNotExists([]byte(e.To().Key()))
		if err := bucket.Put([]byte("attributes"), bits); err != nil {
			return err
		}
		for _, fn := range g.onEdgeChange {
			if err := fn(g, e, false); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *Graph) GetEdge(from graphik.Path, relationship string, to graphik.Path) (graphik.Edge, error) {
	var e = graphik.NewEdge(from, relationship, to, nil)
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(from.Type()))
		if bucket == nil {
			return errors.New("not found")
		}
		bucket = bucket.Bucket([]byte(from.Key()))
		if bucket == nil {
			return errors.New("not found")
		}
		bucket = bucket.Bucket([]byte(relationship))
		if bucket == nil {
			return errors.New("not found")
		}
		bucket = bucket.Bucket([]byte(to.Type()))
		if bucket == nil {
			return errors.New("not found")
		}
		bucket = bucket.Bucket([]byte(to.Key()))
		if bucket == nil {
			return errors.New("not found")
		}
		res := bucket.Get([]byte("attributes"))
		if err := e.Unmarshal(res); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return e, nil
}

func (g *Graph) QueryEdges(q graphik.EdgeQuery) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		if q == nil {
			q = graphik.NewEdgeQuery()
		}
		count := 0
		fromTypes, err := fromTypeBuckets(tx, q.FromType())
		if err != nil {
			return err
		}
		for _, fromType := range fromTypes.buckets {
			fromKeys, err := fromKeyBuckets(fromType, q.FromKey())
			if err != nil {
				return err
			}
			for _, fromKey := range fromKeys.buckets {
				rels, err := relBuckets(fromKey, q.Relationship())
				if err != nil {
					return err
				}
				if rels.key == "attributes" {
					continue
				}
				for _, rel := range rels.buckets {
					toTypes, err := toTypeBuckets(rel, q.ToType())
					if err != nil {
						return err
					}
					for _, toType := range toTypes.buckets {
						toKeys, err := toKeyBuckets(toType, q.ToKey())
						if err != nil {
							return err
						}
						for _, toKey := range toKeys.buckets {
							if q.Limit() != 0 && count >= q.Limit() {
								return nil
							}
							var e = graphik.NewEdge(graphik.NewPath(fromTypes.key, fromKeys.key), rels.key, graphik.NewPath(toTypes.key, toKeys.key), nil)
							res := toKey.Get([]byte("attributes"))
							if len(res) > 0 {
								if err := e.Unmarshal(res); err != nil {
									return err
								}
								if q.Where() != nil {
									if q.Where()(g, e) {
										if err := q.Handler()(g, e); err != nil {
											return err
										}
										count++
									}
								} else {
									if err := q.Handler()(g, e); err != nil {
										return err
									}
									count++
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
}

func (g *Graph) DelEdge(e graphik.Edge) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, fn := range g.edgeConstraints {
			if err := fn(g, e, false); err != nil {
				return err
			}
		}
		bucket := tx.Bucket([]byte(e.From().Type()))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte(e.From().Key()))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte(e.Relationship()))
		if bucket == nil {
			return nil
		}
		bucket = bucket.Bucket([]byte(e.To().Type()))
		if bucket == nil {
			return nil
		}
		if err := bucket.Delete([]byte(e.To().Key())); err != nil {
			return err
		}
		for _, fn := range g.onEdgeChange {
			if err := fn(g, e, true); err != nil {
				return err
			}
		}
		return nil
	})
}

type keyBuckets struct {
	key     string
	buckets []*bbolt.Bucket
}

func fromTypeBuckets(tx *bbolt.Tx, fromType string) (*keyBuckets, error) {
	var buckets []*bbolt.Bucket
	if fromType != "" {
		bucket := tx.Cursor().Bucket().Bucket([]byte(fromType))
		if bucket == nil {
			bucket, _ = tx.Cursor().Bucket().CreateBucketIfNotExists([]byte(fromType))
		}
		buckets = append(buckets, bucket)
	} else {
		if err := tx.Cursor().Bucket().ForEach(func(k, v []byte) error {
			bucket := tx.Cursor().Bucket().Bucket(k)
			if bucket != nil {
				buckets = append(buckets, bucket)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return &keyBuckets{
		key:     fromType,
		buckets: buckets,
	}, nil
}

func fromKeyBuckets(fromType *bbolt.Bucket, fromKey string) (*keyBuckets, error) {
	var buckets []*bbolt.Bucket
	if fromKey != "" {
		bucket := fromType.Bucket([]byte(fromKey))
		if bucket == nil {
			bucket, _ = fromType.CreateBucketIfNotExists([]byte(fromKey))
		}
		buckets = append(buckets, bucket)
	} else {
		if err := fromType.ForEach(func(k, v []byte) error {
			bucket := fromType.Bucket(k)
			if bucket != nil {
				buckets = append(buckets, bucket)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return &keyBuckets{
		key:     fromKey,
		buckets: buckets,
	}, nil
}

func relBuckets(fromKey *bbolt.Bucket, relationship string) (*keyBuckets, error) {
	var buckets []*bbolt.Bucket
	if relationship != "" {
		bucket := fromKey.Bucket([]byte(relationship))
		if bucket == nil {
			bucket, _ = fromKey.CreateBucketIfNotExists([]byte(relationship))
		}
		buckets = append(buckets, bucket)
	} else {
		if err := fromKey.ForEach(func(k, v []byte) error {
			bucket := fromKey.Bucket(k)
			if bucket != nil {
				buckets = append(buckets, bucket)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return &keyBuckets{
		key:     relationship,
		buckets: buckets,
	}, nil
}

func toTypeBuckets(relationship *bbolt.Bucket, toType string) (*keyBuckets, error) {
	var buckets []*bbolt.Bucket
	if toType != "" {
		bucket := relationship.Bucket([]byte(toType))
		if bucket == nil {
			bucket, _ = relationship.CreateBucketIfNotExists([]byte(toType))
		}
		buckets = append(buckets, bucket)
	} else {
		if err := relationship.ForEach(func(k, v []byte) error {
			bucket := relationship.Bucket(k)
			if bucket != nil {
				buckets = append(buckets, bucket)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return &keyBuckets{
		key:     toType,
		buckets: buckets,
	}, nil
}

func toKeyBuckets(toType *bbolt.Bucket, toKey string) (*keyBuckets, error) {
	var buckets []*bbolt.Bucket
	if toKey != "" {
		bucket := toType.Bucket([]byte(toKey))
		if bucket == nil {
			bucket, _ = toType.CreateBucketIfNotExists([]byte(toKey))
		}
		buckets = append(buckets, bucket)
	} else {
		if err := toType.ForEach(func(k, v []byte) error {
			bucket := toType.Bucket(k)
			if bucket != nil {
				buckets = append(buckets, bucket)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return &keyBuckets{
		key:     toKey,
		buckets: buckets,
	}, nil
}
