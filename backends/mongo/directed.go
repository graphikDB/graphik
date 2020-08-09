package mongo

import (
	"context"
	"github.com/autom8ter/graphik"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Graph struct {
	edges *mongo.Database
	nodes *mongo.Database
	uri   string
}

// Open returns a Graph.
func Open(uri string) graphik.GraphOpenerFunc {
	return func() (graphik.Graph, error) {
		client, err := mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}
		err = client.Ping(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Graph{
			uri:   uri,
			nodes: client.Database("nodes"),
			edges: client.Database("edges"),
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

	_, err := g.nodes.Collection(n.Type()).ReplaceOne(ctx, bson.D{{
		Key:   "_id",
		Value: n.PathString()},
	}, vals, opts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Graph) nodeTypes(ctx context.Context) ([]string, error) {
	return g.nodes.ListCollectionNames(ctx, nil)
}

func (g *Graph) edgeRelationships(ctx context.Context) ([]string, error) {
	return g.edges.ListCollectionNames(ctx, nil)
}

func (g *Graph) QueryNodes(ctx context.Context, query graphik.NodeQuery) error {
	if query == nil {
		query = graphik.NewNodeQuery()
	}
	if query.Closer() != nil {
		defer query.Closer()
	}
	filter := []bson.E{}
	if query.Type() != "" {
		filter = append(filter, bson.E{
			Key:   "type",
			Value: query.Type(),
		})
	}
	if query.Key() != "" {
		filter = append(filter, bson.E{
			Key:   "key",
			Value: query.Key(),
		})
	}
	var paths []string
	if query.Type() == "" {
		types, err := g.nodeTypes(ctx)
		if err != nil {
			return err
		}
		paths = append(paths, types...)
	} else {
		paths = []string{query.Type()}
	}
	lim := query.Limit()
	if lim == 0 {
		lim = 1000
	}
	for _, collection := range paths {
		cursor, err := g.nodes.Collection(collection).Find(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		counter := 0
		for cursor.Next(ctx) && counter < lim {
			raw := cursor.Current
			var n = graphik.NewNode(graphik.NewPath(collection, ""))
			if err := n.Unmarshal(raw); err != nil {
				return err
			}
			if query.Where() != nil {
				if query.Where()(g, n) {
					if err := query.Handler()(g, n); err != nil {
						return err
					}
					counter++
				}
			} else {
				if err := query.Handler()(g, n); err != nil {
					return err
				}
				counter++
			}
		}
	}
	return nil

}

func (g *Graph) DelNode(ctx context.Context, path graphik.Path) error {
	if err := g.nodes.Collection(path.Type()).FindOneAndDelete(ctx, bson.D{{
		Key:   "_id",
		Value: path.Key()},
	}).Err(); err != nil {
		return err
	}
	return nil
}

func (g *Graph) GetNode(ctx context.Context, path graphik.Path) (graphik.Node, error) {
	res := g.nodes.Collection(path.Type()).FindOne(ctx, bson.D{{
		Key:   "_id",
		Value: path.PathString()},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	bits, err := res.DecodeBytes()
	if err != nil {
		return nil, err
	}
	n := graphik.NewNode(path)
	if err := n.Unmarshal(bits); err != nil {
		return nil, err
	}
	return n, nil
}

func (g *Graph) AddEdge(ctx context.Context, e graphik.Edge) error {
	opts := options.Replace().SetUpsert(true)
	vals := bson.M{}
	e.Range(func(k string, v interface{}) bool {
		vals[k] = v
		return true
	})
	_, err := g.edges.Collection(e.Relationship()).ReplaceOne(ctx, bson.D{{
		Key:   "_id",
		Value: e.PathString()},
	}, vals, opts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Graph) GetEdge(ctx context.Context, path graphik.EdgePath) (graphik.Edge, error) {
	res := g.edges.Collection(path.Relationship()).FindOne(ctx, bson.D{{
		Key:   "_id",
		Value: path.PathString()},
	})
	if res.Err() != nil {
		return nil, res.Err()
	}
	bits, err := res.DecodeBytes()
	if err != nil {
		return nil, err
	}
	e := graphik.NewEdge(path)
	if err := e.Unmarshal(bits); err != nil {
		return nil, err
	}
	return e, nil
}

func (g *Graph) QueryEdges(ctx context.Context, query graphik.EdgeQuery) error {
	if query == nil {
		query = graphik.NewEdgeQuery()
	}
	if query.Closer() != nil {
		defer query.Closer()()
	}
	filter := []bson.E{}
	if query.FromType() != "" {
		filter = append(filter, bson.E{
			Key:   "fromType",
			Value: query.FromType(),
		})
	}
	if query.FromKey() != "" {
		filter = append(filter, bson.E{
			Key:   "fromKey",
			Value: query.FromKey(),
		})
	}
	if query.ToType() != "" {
		filter = append(filter, bson.E{
			Key:   "toType",
			Value: query.ToType(),
		})
	}
	if query.ToKey() != "" {
		filter = append(filter, bson.E{
			Key:   "toKey",
			Value: query.ToKey(),
		})
	}
	var collections []string
	if query.Relationship() == "" {
		relationships, err := g.edgeRelationships(ctx)
		if err != nil {
			return err
		}
		collections = append(collections, relationships...)
	} else {
		collections = []string{query.Relationship()}
	}
	lim := query.Limit()
	if lim == 0 {
		lim = 1000
	}

	for _, collection := range collections {
		cursor, err := g.edges.Collection(collection).Find(ctx, filter)
		if err != nil {
			return errors.Wrap(err, "failed to find cursor")
		}
		defer cursor.Close(ctx)
		//counter := 0
		if err := cursor.Err(); err != nil {
			return errors.Wrap(err, "cursor error")
		}
		for cursor.Next(ctx) {
			raw := cursor.Current
			var e = graphik.NewEdge(graphik.NewEdgePath(
				graphik.NewPath(query.FromType(), query.FromKey()),
				collection,
				graphik.NewPath(query.ToType(), query.ToKey()),
			))
			if err := e.Unmarshal(raw); err != nil {
				return errors.Wrap(err, "failed to marshal edge")
			}
			if query.Where() != nil {
				if query.Where()(g, e) {
					if err := query.Handler()(g, e); err != nil {
						return err
					}
				}
			} else {
				if err := query.Handler()(g, e); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return errors.Wrap(err, "cursor error")
		}
	}
	return nil
}

func (g *Graph) DelEdge(ctx context.Context, e graphik.Edge) error {
	res := g.edges.Collection(e.Relationship()).FindOne(ctx, bson.D{{
		Key:   "_id",
		Value: e.PathString()},
	})
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (g *Graph) Close(ctx context.Context) error {
	return g.nodes.Client().Disconnect(ctx)
}
