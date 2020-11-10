package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/values"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
)

func (a *Runtime) ToContext(ctx context.Context, payload map[string]interface{}) (context.Context, error) {
	path := &apipb.Path{
		Type: identityType,
		Id:   payload[idClaim].(string),
	}
	var err error
	n, ok := a.graph.GetNode(path)
	if !ok {
		strct, _ := structpb.NewStruct(payload)
		n, err = a.CreateNode(&apipb.Node{
			Type:       path.Type,
			Id:         path.Id,
			Attributes: strct,
		})
		if err != nil {
			return nil, err
		}
	}
	return context.WithValue(ctx, authCtxKey, n), nil
}

func (s *Runtime) NodeContext(ctx context.Context) values.Values {
	val, ok := ctx.Value(authCtxKey).(values.Values)
	if ok {
		return val
	}
	return nil
}
