package graph

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GraphStore) setNode(ctx context.Context, tx *bbolt.Tx, node *apipb.Node) (*apipb.Node, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if node.GetPath() == nil {
		node.Path = &apipb.Path{}
	}
	if node.GetPath().Gid == "" {
		node.Path.Gid = uuid.New().String()
	}
	identity := g.getIdentity(ctx)
	if node.Metadata == nil {
		node.Metadata = &apipb.Metadata{}
	}
	if node.GetMetadata().GetUpdatedBy() == nil {
		identity.GetMetadata().UpdatedBy = identity.GetPath()
	}
	if node.GetMetadata().GetCreatedAt() == nil {
		node.GetMetadata().CreatedAt = timestamppb.Now()
	}
	if node.GetMetadata().GetUpdatedAt() == nil {
		node.GetMetadata().UpdatedAt = timestamppb.Now()
	}
	node.GetMetadata().Version += 1
	bits, err := proto.Marshal(node)
	if err != nil {
		return nil, err
	}
	nodeBucket := tx.Bucket(dbNodes)
	bucket := nodeBucket.Bucket([]byte(node.GetPath().GetGtype()))
	if bucket == nil {
		bucket, err = nodeBucket.CreateBucketIfNotExists([]byte(node.GetPath().GetGtype()))
		if err != nil {
			return nil, err
		}
	}
	if err := bucket.Put([]byte(node.GetPath().GetGid()), bits); err != nil {
		return nil, err
	}
	return node, nil
}

func (g *GraphStore) setNodes(ctx context.Context, nodes ...*apipb.Node) (*apipb.Nodes, error) {
	var nds = &apipb.Nodes{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, node := range nodes {
			n, err := g.setNode(ctx, tx, node)
			if err != nil {
				return err
			}
			nds.Nodes = append(nds.Nodes, n)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return nds, nil
}

func (g *GraphStore) setEdge(ctx context.Context, tx *bbolt.Tx, edge *apipb.Edge) (*apipb.Edge, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	nodeBucket := tx.Bucket(dbNodes)
	{
		fromBucket := nodeBucket.Bucket([]byte(edge.From.Gtype))
		if fromBucket == nil {
			return nil, errors.Errorf("from node %s does not exist", edge.From.String())
		}
		val := fromBucket.Get([]byte(edge.From.Gid))
		if val == nil {
			return nil, errors.Errorf("from node %s does not exist", edge.From.String())
		}
	}
	{
		toBucket := nodeBucket.Bucket([]byte(edge.To.Gtype))
		if toBucket == nil {
			return nil, errors.Errorf("to node %s does not exist", edge.To.String())
		}
		val := toBucket.Get([]byte(edge.To.Gid))
		if val == nil {
			return nil, errors.Errorf("to node %s does not exist", edge.To.String())
		}
	}
	if edge.GetPath() == nil {
		edge.Path = &apipb.Path{}
	}
	if edge.GetPath().Gid == "" {
		edge.Path.Gid = uuid.New().String()
	}
	identity := g.getIdentity(ctx)
	if edge.Metadata == nil {
		edge.Metadata = &apipb.Metadata{}
	}
	now := timestamppb.Now()
	if edge.GetMetadata().GetCreatedBy() == nil {
		identity.GetMetadata().CreatedBy = identity.GetPath()
	}
	if edge.GetMetadata().GetUpdatedBy() == nil {
		identity.GetMetadata().UpdatedBy = identity.GetPath()
	}
	if edge.GetMetadata().GetCreatedAt() == nil {
		edge.GetMetadata().CreatedAt = now
	}
	if edge.GetMetadata().GetUpdatedAt() == nil {
		edge.GetMetadata().UpdatedAt = now
	}
	edge.GetMetadata().Version += 1
	bits, err := proto.Marshal(edge)
	if err != nil {
		return nil, err
	}
	edgeBucket := tx.Bucket(dbEdges)
	edgeBucket = edgeBucket.Bucket([]byte(edge.GetPath().GetGtype()))
	if edgeBucket == nil {
		edgeBucket, err = edgeBucket.CreateBucketIfNotExists([]byte(edge.GetPath().GetGtype()))
		if err != nil {
			return nil, err
		}
	}
	if err := edgeBucket.Put([]byte(edge.GetPath().GetGid()), bits); err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.edgesFrom[edge.GetFrom().String()] = append(g.edgesFrom[edge.GetFrom().String()], edge.GetPath())
	g.edgesTo[edge.GetTo().String()] = append(g.edgesTo[edge.GetTo().String()], edge.GetPath())
	if !edge.Directed {
		g.edgesTo[edge.GetFrom().String()] = append(g.edgesTo[edge.GetFrom().String()], edge.GetPath())
		g.edgesFrom[edge.GetTo().String()] = append(g.edgesFrom[edge.GetTo().String()], edge.GetPath())
	}
	sortPaths(g.edgesFrom[edge.GetFrom().String()])
	sortPaths(g.edgesTo[edge.GetTo().String()])
	return edge, nil
}

func (g *GraphStore) setEdges(ctx context.Context, edges ...*apipb.Edge) (*apipb.Edges, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var edgs = &apipb.Edges{}
	if err := g.db.Update(func(tx *bbolt.Tx) error {
		for _, edge := range edges {
			e, err := g.setEdge(ctx, tx, edge)
			if err != nil {
				return err
			}
			edgs.Edges = append(edgs.Edges, e)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	edgs.Sort()
	return edgs, nil
}

func (g *GraphStore) getNode(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) (*apipb.Node, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	identity := g.getIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get identity")
	}
	var node apipb.Node
	bucket := tx.Bucket(dbNodes).Bucket([]byte(path.Gtype))
	if bucket == nil {
		return nil, ErrNotFound
	}
	bits := bucket.Get([]byte(path.Gid))
	if err := proto.Unmarshal(bits, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func (g *GraphStore) getEdge(ctx context.Context, tx *bbolt.Tx, path *apipb.Path) (*apipb.Edge, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var edge apipb.Edge
	bucket := tx.Bucket(dbEdges).Bucket([]byte(path.Gtype))
	if bucket == nil {
		return nil, ErrNotFound
	}
	bits := bucket.Get([]byte(path.Gid))
	if err := proto.Unmarshal(bits, &edge); err != nil {
		return nil, err
	}
	return &edge, nil
}
