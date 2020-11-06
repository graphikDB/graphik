package runtime

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		newNode, err := a.CreateNodes(&apipb.Nodes{
			Nodes: []*apipb.Node{
				{
					Path:       path,
					Attributes: n.Attributes,
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

func (r *Runtime) JWTMiddleware() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := r.jwks.VerifyJWT(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		if exp, ok := payload["exp"].(int64); ok {
			if exp < time.Now().Unix() {
				return nil, status.Errorf(codes.Unauthenticated, "token expired")
			}
		}
		if exp, ok := payload["exp"].(int); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, status.Errorf(codes.Unauthenticated, "token expired")
			}
		}
		ctx, err = r.toContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		return ctx, nil
	}
}
