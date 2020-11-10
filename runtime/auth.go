package runtime

import (
	"context"
	"github.com/autom8ter/graphik/graph"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
)

func (a *Runtime) ToContext(ctx context.Context, payload map[string]interface{}) (context.Context, error) {
	path := graph.FormPath(identityType, payload[idClaim].(string))
	n, ok := a.graph.Nodes().Get(path)
	if !ok {
		values := map[string]interface{}{
			"path": path,
		}
		for k, v := range payload {
			values[k] = v
		}
		newNode, err := a.CreateNodes(graph.ValueSet{
			values,
		})
		if err != nil {
			return nil, err
		}
		n = newNode[0]
	}
	return context.WithValue(ctx, authCtxKey, n), nil
}

func (s *Runtime) NodeContext(ctx context.Context) graph.Values {
	val, ok := ctx.Value(authCtxKey).(graph.Values)
	if ok {
		return val
	}
	return nil
}
