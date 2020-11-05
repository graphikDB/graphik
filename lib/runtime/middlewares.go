package runtime

import (
	"context"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (a *Runtime) AuthMiddleware() grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := a.opts.jwks.VerifyJWT(token)
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
		ctx, err = a.toContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		return ctx, nil
	}
}
