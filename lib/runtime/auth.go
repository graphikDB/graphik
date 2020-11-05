package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func (a *Runtime) toContext(ctx context.Context, payload map[string]interface{}) context.Context {
	path := &apipb.Path{
		Type: "user",
		ID:   payload["sub"].(string),
	}
	n, ok := a.nodes.Get(path)
	if !ok {
		n = a.nodes.Set(&apipb.Node{
			Path:       path,
			Attributes: n.Attributes,
		})
	}
	return context.WithValue(ctx, authCtxKey, n)
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
