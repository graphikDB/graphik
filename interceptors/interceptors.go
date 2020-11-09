package interceptors

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	"github.com/autom8ter/graphik/runtime"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func UnaryAuth(runtime *runtime.Runtime) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := runtime.Auth().VerifyJWT(token)
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
		ctx, err = runtime.ToContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		request := &apipb.UserIntercept{
			RequestPath: info.FullMethod,
			User:        runtime.NodeContext(ctx),
			Request:     lang.ToStruct(lang.ToMap(req)),
		}
		pass, err := runtime.Auth().Authorize(request)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		if !pass {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		return handler(ctx, req)
	}
}

func StreamAuth(runtime *runtime.Runtime) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := grpc_auth.AuthFromMD(ss.Context(), "Bearer")
		if err != nil {
			return err
		}
		payload, err := runtime.Auth().VerifyJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}
		if payload["exp"].(int64) < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}

		ctx, err := runtime.ToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		request := &apipb.UserIntercept{
			RequestPath: info.FullMethod,
			User:        runtime.NodeContext(ctx),
			Request:     lang.ToStruct(lang.ToMap(srv)),
		}
		pass, err := runtime.Auth().Authorize(request)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		if !pass {
			return status.Error(codes.PermissionDenied, "permission denied")
		}

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}
