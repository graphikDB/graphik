package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"time"
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
		any, _ := ptypes.MarshalAny(&apipb.Node{
			Path:       path,
			Attributes: n.Attributes,
		})
		val, err := a.Execute(&apipb.Command{
			Op:  apipb.Op_CREATE_NODE,
			Val: any,
			Timestamp: &timestamp.Timestamp{
				Seconds: time.Now().Unix(),
			},
		})
		if err != nil {
			return nil, err
		}
		n = val.(*apipb.Node)
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
