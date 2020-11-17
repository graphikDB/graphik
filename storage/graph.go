package storage

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"go.etcd.io/bbolt"
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

func (g *GraphStore) GetEdge(path *apipb.Path) (*apipb.Edge, error) {
	var edge apipb.Edge
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges)
		bucket, err := bucket.CreateBucketIfNotExists([]byte(path.Gtype))
		if err != nil {
			return err
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

func (g *GraphStore) RangeEdges(gType string, fn func(e *apipb.Edge) bool) error {
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges).Bucket([]byte(gType))
		return bucket.ForEach(func(k, v []byte) error {
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

func (g *GraphStore) GetNode(path *apipb.Path) (*apipb.Node, error) {
	var node apipb.Node
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes)
		bucket, err := bucket.CreateBucketIfNotExists([]byte(path.Gtype))
		if err != nil {
			return err
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

func (g *GraphStore) RangeNodes(gType string, fn func(n *apipb.Node) bool) error {
	if err := g.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes).Bucket([]byte(gType))
		return bucket.ForEach(func(k, v []byte) error {
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

func (g *GraphStore) SetNode(node *apipb.Node) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		bits, err := proto.Marshal(node)
		if err != nil {
			return err
		}
		bucket := tx.Bucket(dbNodes)
		bucket, err = bucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(node.GetPath().GetGid()), bits)
	})
}

func (g *GraphStore) SetEdge(edge *apipb.Edge) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		bits, err := proto.Marshal(edge)
		if err != nil {
			return err
		}
		bucket := tx.Bucket(dbEdges)
		bucket, err = bucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(edge.GetPath().GetGid()), bits)
	})
}

func (g *GraphStore) DelEdges(paths ...*apipb.Path) error {
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

func (g *GraphStore) DelNodes(paths ...*apipb.Path) error {
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

func (g *GraphStore) DelNodeType(typ string) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes)
		return bucket.DeleteBucket([]byte(typ))
	})
}

func (g *GraphStore) DelEdgeType(typ string) error {
	return g.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEdges)
		return bucket.DeleteBucket([]byte(typ))
	})
}
