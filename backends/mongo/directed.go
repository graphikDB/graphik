package mongo

import (
	"context"
	"github.com/autom8ter/graphik"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Graph struct {
	db  *mongo.Database
	uri string
}

// Open returns a Graph.
func Open(uri string) graphik.GraphOpenerFunc {
	return func() (graphik.Graph, error) {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			return nil, err
		}
		return &Graph{
			uri: uri,
			db:  client.Database("graphik"),
		}, nil
	}
}

func (g *Graph) implements() graphik.Graph {
	return g
}

func (g *Graph) AddNode(ctx context.Context, n graphik.Node) error {
	opts := options.Replace().SetUpsert(true)
	vals := bson.M{}
	n.Range(func(k string, v interface{}) bool {
		vals[k] = v
		return true
	})
	_, err := g.db.Collection(n.Type()).ReplaceOne(ctx, bson.D{{
		Key:   "_id",
		Value: n.Key()},
	}, vals, opts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Graph) QueryNodes(ctx context.Context, query graphik.NodeQuery) error {
	panic("implement me")
}

func (g *Graph) DelNode(ctx context.Context, path graphik.Path) error {
	if err := g.db.Collection(path.Type()).FindOneAndDelete(ctx, bson.D{{
		Key:   "_id",
		Value: path.Key()},
	}).Err(); err != nil {
		return err
	}
	return nil
}

func (g *Graph) GetNode(ctx context.Context, path graphik.Path) (graphik.Node, error) {
	res := g.db.Collection(path.Type()).FindOne(ctx, bson.D{{
		Key:   "_id",
		Value: path.Key()},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	bits, err := res.DecodeBytes()
	if err != nil {
		return nil, err
	}
	n := graphik.NewNode(path, nil)
	if err := n.Unmarshal(bits); err != nil {
		return nil, err
	}
	return n, nil
}

func (g *Graph) AddEdge(ctx context.Context, e graphik.Edge) error {
	panic("implement me")
}

func (g *Graph) GetEdge(ctx context.Context, from graphik.Path, relationship string, to graphik.Path) (graphik.Edge, error) {
	panic("implement me")
}

func (g *Graph) QueryEdges(ctx context.Context, query graphik.EdgeQuery) error {
	panic("implement me")
}

func (g *Graph) DelEdge(ctx context.Context, e graphik.Edge) error {
	panic("implement me")
}

func (g *Graph) Close(ctx context.Context) error {
	return g.db.Client().Disconnect(ctx)
}
