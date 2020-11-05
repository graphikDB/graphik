package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func (a *Runtime) toContext(ctx context.Context, payload map[string]interface{}) (context.Context, error) {
	path := &apipb.Path{
		Type: "user",
		ID:   payload["sub"].(string),
	}
	n, ok := a.nodes.Get(path)
	if !ok {
		newNode, err := a.CreateNode(&apipb.Node{
			Path:       path,
			Attributes: n.Attributes,
		})
		if err != nil {
			return nil, err
		}
		n = newNode
	}
	return context.WithValue(ctx, authCtxKey, n), nil
}

func (s *Runtime) GetNode(ctx context.Context) *apipb.Node {
	val, ok := ctx.Value(authCtxKey).(*apipb.Node)
	if ok {
		return val
	}
	val2, ok := ctx.Value(authCtxKey).(apipb.Node)
	if ok {
		return &val2
	}
	return nil
}
