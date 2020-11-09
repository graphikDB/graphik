package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
)

func (a *Runtime) ToContext(ctx context.Context, payload map[string]interface{}) (context.Context, error) {
	path := lang.FormPath(identityType, payload[idClaim].(string))
	n, ok := a.graph.Nodes().Get(path)
	if !ok {
		newNode, err := a.CreateNodes(&apipb.Nodes{
			Nodes: []*apipb.Node{
				{
					Path:       path,
					Attributes: lang.ToStruct(payload),
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
