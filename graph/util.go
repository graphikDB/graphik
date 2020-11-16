package graph

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/dgraph-io/badger/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

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
	return errors.Errorf("%s.%s does not exist", path.Gtype, path.Gid)
}

func (g *Graph) getEdge(path *apipb.Path) (*apipb.Edge, error) {
	var edge apipb.Edge
	if err := g.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("edges/%s/%s", path.GetGtype(), path.GetGid())))
		if err != nil {
			return err
		}
		valCopy, err := item.ValueCopy(nil)
		if err := proto.Unmarshal(valCopy, &edge); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &edge, nil
}

func (g *Graph) getNode(path *apipb.Path) (*apipb.Node, error) {
	var node apipb.Node
	if err := g.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("nodes/%s/%s", path.GetGtype(), path.GetGid())))
		if err != nil {
			return err
		}
		valCopy, err := item.ValueCopy(nil)
		if err := proto.Unmarshal(valCopy, &node); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &node, nil
}

func (g *Graph) setNode(node *apipb.Node) error {
	return g.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("nodes/%s/%s", node.GetPath().GetGtype(), node.GetPath().GetGid()))
		bits, err := proto.Marshal(node)
		if err != nil {
			return err
		}
		if err := txn.Set(key, bits); err != nil {
			return err
		}
		return nil
	})
}

func (g *Graph) setEdge(edge *apipb.Edge) error {
	return g.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("edges/%s/%s", edge.GetPath().GetGtype(), edge.GetPath().GetGid()))
		bits, err := proto.Marshal(edge)
		if err != nil {
			return err
		}
		if err := txn.Set(key, bits); err != nil {
			return err
		}
		return nil
	})
}

func (g *Graph) delEdge(paths ...*apipb.Path) error {
	return g.db.Update(func(txn *badger.Txn) error {
		for _, path := range paths {
			if err := txn.Delete([]byte(fmt.Sprintf("edges/%s/%s", path.GetGtype(), path.GetGid()))); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *Graph) delNode(paths ...*apipb.Path) error {
	return g.db.Update(func(txn *badger.Txn) error {
		for _, path := range paths {
			if err := txn.Delete([]byte(fmt.Sprintf("nodes/%s/%s", path.GetGtype(), path.GetGid()))); err != nil {
				return err
			}
		}
		return nil
	})
}
