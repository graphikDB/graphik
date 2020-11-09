package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
)

const (
	authCtxKey = "x-graphik-auth-ctx"
)

func (a *Runtime) ToContext(ctx context.Context, payload map[string]interface{}) (context.Context, error) {
	path := &apipb.Path{
		Type: "user",
		ID:   payload["sub"].(string),
	}
	n, ok := a.nodes.Get(path)
	if !ok {
		newNode, err := a.CreateNodes(&apipb.Nodes{
			Nodes: []*apipb.Node{
				{
					Path:       path,
					Attributes: apipb.ToStruct(payload),
				},
			},
		})
		if err != nil {
			return nil, err
		}
		n = newNode.Nodes[0]
	}
	return context.WithValue(ctx, authCtxKey, n), nil
}

func (s *Runtime) NodeContext(ctx context.Context) *apipb.Node {
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
