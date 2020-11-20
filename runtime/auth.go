package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
)

func (a *Runtime) ToContext(ctx context.Context, payload map[string]interface{}) (context.Context, *apipb.Node, error) {
	var err error
	n, err := a.graph.GetNode(ctx, &apipb.Path{
		Gtype: identityType,
		Gid:   payload[idClaim].(string),
	})
	if err != nil || n == nil {
		strct, _ := structpb.NewStruct(payload)
		n, err = a.CreateNode(ctx, &apipb.NodeConstructor{
			Path: &apipb.Path{
				Gtype: identityType,
				Gid:   payload[idClaim].(string),
			},
			Attributes: strct,
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return context.WithValue(ctx, authCtxKey, n), n, nil
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

func (r *Runtime) MethodContext(ctx context.Context) string {
	val, ok := ctx.Value(methodCtxKey).(string)
	if ok {
		return val
	}
	return ""
}
