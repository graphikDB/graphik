package storage

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/bbolt"
	"sort"
)

type GraphStore struct {
	// db is the underlying handle to the db.
	db *bbolt.DB
	// The path to the Bolt database file
	path string
}

// NewGraphStore takes a file path and returns a connected Raft backend.
func NewGraphStore(path string) (*GraphStore, error) {
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
	return &GraphStore{
		db:   handle,
		path: path,
	}, nil
}

// Close is used to gracefully close the DB connection.
func (b *GraphStore) Close() error {
	return b.db.Close()
}

func (g *GraphStore) GetEdge(ctx context.Context, path *apipb.Path) (*apipb.Edge, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var edge apipb.Edge
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bits := tx.Bucket(dbEdges).Bucket([]byte(path.Gtype)).Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, &edge); err != nil {
			return err
		}
		return nil
	}); err != nil && err != DONE {
		return nil, err
	}
	return &edge, nil
}

func (g *GraphStore) RangeEdges(ctx context.Context, gType string, fn func(e *apipb.Edge) bool) error {
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
			if err := g.RangeEdges(ctx, edgeType, fn); err != nil {
				return err
			}
		}
		return nil
	}
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dbEdges).Bucket([]byte(gType)).ForEach(func(k, v []byte) error {
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
	var node apipb.Node
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bits := tx.Bucket(dbNodes).Bucket([]byte(path.Gtype)).Get([]byte(path.Gid))
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
		return tx.Bucket(dbNodes).Bucket([]byte(gType)).ForEach(func(k, v []byte) error {
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

func (g *GraphStore) SetNode(ctx context.Context, node *apipb.Node) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
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
		return bucket.Put([]byte(node.GetPath().GetGid()), bits)
	})
}

func (g *GraphStore) SetNodes(ctx context.Context, nodes ...*apipb.Node) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, node := range nodes {
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
	})
}

func (g *GraphStore) SetEdge(ctx context.Context, edge *apipb.Edge) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
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
		return bucket.Put([]byte(edge.GetPath().GetGid()), bits)
	})
}

func (g *GraphStore) SetEdges(ctx context.Context, edges ...*apipb.Edge) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
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
		}
		return nil
	})
}

func (g *GraphStore) DelEdges(ctx context.Context, paths ...*apipb.Path) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, p := range paths {
			bucket := tx.Bucket(dbEdges)
			bucket = bucket.Bucket([]byte(p.GetGtype()))
			if bucket == nil {
				return nil
			}
			return bucket.Delete([]byte(p.GetGid()))
		}
		return nil
	})
}

func (g *GraphStore) DelNodes(ctx context.Context, paths ...*apipb.Path) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, p := range paths {
			bucket := tx.Bucket(dbNodes)
			bucket = bucket.Bucket([]byte(p.GetGtype()))
			if bucket == nil {
				return nil
			}
			return bucket.Delete([]byte(p.GetGid()))
		}
		return nil
	})
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
