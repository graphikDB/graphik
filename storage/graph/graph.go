package graph

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/sortable"
	"github.com/autom8ter/graphik/storage"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"sort"
	"sync"
)

type GraphStore struct {
	// db is the underlying handle to the db.
	db *bbolt.DB
	// The path to the Bolt database file
	path      string
	mu        sync.RWMutex
	edgesTo   map[string][]*apipb.Path
	edgesFrom map[string][]*apipb.Path
}

// NewGraphStore takes a file path and returns a connected Raft backend.
func NewGraphStore(path string) (*GraphStore, error) {
	handle, err := bbolt.Open(path, storage.DbFileMode, nil)
	if err != nil {
		return nil, err
	}
	tx, err := handle.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(storage.DbNodes); err != nil {
		return nil, err
	}
	if _, err := tx.CreateBucketIfNotExists(storage.DbEdges); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &GraphStore{
		db:        handle,
		path:      path,
		mu:        sync.RWMutex{},
		edgesTo:   map[string][]*apipb.Path{},
		edgesFrom: map[string][]*apipb.Path{},
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
		bucket := tx.Bucket(storage.DbEdges).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return storage.ErrNotFound
		}
		bits := bucket.Get(storage.Uint64ToBytes(path.Gid))
		if err := proto.Unmarshal(bits, &edge); err != nil {
			return err
		}
		return nil
	}); err != nil && err != storage.DONE {
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
		bucket := tx.Bucket(storage.DbEdges).Bucket([]byte(gType))
		if bucket == nil {
			return storage.ErrNotFound
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
				return storage.DONE
			}
		}
		return nil
	}); err != nil && err != storage.DONE {
		return err
	}
	return nil
}

func (g *GraphStore) RangeEdges() {

}

func (g *GraphStore) GetNode(ctx context.Context, path *apipb.Path) (*apipb.Node, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var node apipb.Node
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(storage.DbNodes).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return storage.ErrNotFound
		}
		bits := bucket.Get(storage.Uint64ToBytes(path.Gid))
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
		bucket := tx.Bucket(storage.DbNodes).Bucket([]byte(gType))
		if bucket == nil {
			return storage.ErrNotFound
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
				return storage.DONE
			}
			return nil
		})
	}); err != nil && err != storage.DONE {
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
		nodeBucket := tx.Bucket(storage.DbNodes)
		bucket := nodeBucket.Bucket([]byte(node.GetPath().GetGtype()))
		if bucket == nil {
			bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
			if err != nil {
				return err
			}
		}
		return bucket.Put(storage.Uint64ToBytes(node.GetPath().GetGid()), bits)
	})
}

func (g *GraphStore) SetNodes(ctx context.Context, nodes ...*apipb.Node) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		for _, node := range nodes {
			if err := ctx.Err(); err != nil {
				return err
			}
			bits, err := proto.Marshal(node)
			if err != nil {
				return err
			}
			nodeBucket := tx.Bucket(storage.DbNodes)
			bucket := nodeBucket.Bucket([]byte(node.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if err := bucket.Put(storage.Uint64ToBytes(node.GetPath().GetGid()), bits); err != nil {
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
		edgeBucket := tx.Bucket(storage.DbEdges)
		bucket := edgeBucket.Bucket([]byte(edge.GetPath().GetGtype()))
		if bucket == nil {
			bucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
			if err != nil {
				return err
			}
		}
		return bucket.Put(storage.Uint64ToBytes(edge.GetPath().GetGid()), bits)
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
			edgeBucket := tx.Bucket(storage.DbEdges)
			bucket := edgeBucket.Bucket([]byte(edge.GetPath().GetGtype()))
			if bucket == nil {
				bucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
				if err != nil {
					return err
				}
			}
			if err := bucket.Put(storage.Uint64ToBytes(edge.GetPath().GetGid()), bits); err != nil {
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
			bucket := tx.Bucket(storage.DbEdges)
			bucket = bucket.Bucket([]byte(p.GetGtype()))
			if bucket == nil {
				return storage.ErrNotFound
			}
			return bucket.Delete(storage.Uint64ToBytes(p.GetGid()))
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
			bucket := tx.Bucket(storage.DbNodes)
			bucket = bucket.Bucket([]byte(p.GetGtype()))
			if bucket == nil {
				return storage.ErrNotFound
			}
			return bucket.Delete(storage.Uint64ToBytes(p.GetGid()))
		}
		return nil
	})
}

func (g *GraphStore) DelNodeType(ctx context.Context, typ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(storage.DbNodes)
		return bucket.DeleteBucket([]byte(typ))
	})
}

func (g *GraphStore) DelEdgeType(ctx context.Context, typ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(storage.DbEdges)
		return bucket.DeleteBucket([]byte(typ))
	})
}

func (g *GraphStore) EdgeTypes(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var types []string
	if err := g.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(storage.DbEdges).ForEach(func(name []byte, _ []byte) error {
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
		return tx.Bucket(storage.DbNodes).ForEach(func(name []byte, _ []byte) error {
			types = append(types, string(name))
			return nil
		})
	}); err != nil {
		return nil, err
	}
	sort.Strings(types)
	return types, nil
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

func noExist(path *apipb.Path) error {
	return errors.Errorf("%s.%v does not exist", path.Gtype, path.Gid)
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
