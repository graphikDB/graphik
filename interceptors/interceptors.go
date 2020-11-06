package interceptors

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
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
		n := runtime.NodeContext(ctx)
		request := &apipb.RequestIntercept{
			RequestPath: info.FullMethod,
			User:        n,
		}
		populate(req, request)
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
		if exp, ok := payload["exp"].(int64); ok {
			if exp < time.Now().Unix() {
				return status.Errorf(codes.Unauthenticated, "token expired")
			}
		}
		if exp, ok := payload["exp"].(int); ok {
			if int64(exp) < time.Now().Unix() {
				return status.Errorf(codes.Unauthenticated, "token expired")
			}
		}

		ctx, err := runtime.ToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		n := runtime.NodeContext(ctx)

		request := &apipb.RequestIntercept{
			RequestPath: info.FullMethod,
			User:        n,
		}
		populate(srv, request)
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

func populate(req interface{}, intercept *apipb.RequestIntercept) {
	switch r := req.(type) {
	case *apipb.SetAuthRequest:
		intercept.Request = &apipb.RequestIntercept_SetAuth{SetAuth: r}
	case *apipb.SearchNodesRequest:
		intercept.Request = &apipb.RequestIntercept_SearchNodes{
			SearchNodes: r,
		}
	case *apipb.SearchEdgesRequest:
		intercept.Request = &apipb.RequestIntercept_SearchEdges{
			SearchEdges: r,
		}
	case *apipb.PatchNodesRequest:
		intercept.Request = &apipb.RequestIntercept_PatchNodes{
			PatchNodes: r,
		}
	case *apipb.PatchEdgesRequest:
		intercept.Request = &apipb.RequestIntercept_PatchEdges{
			PatchEdges: r,
		}
	case *apipb.EdgesToRequest:
		intercept.Request = &apipb.RequestIntercept_EdgesTo{
			EdgesTo: r,
		}
	case *apipb.EdgesFromRequest:
		intercept.Request = &apipb.RequestIntercept_EdgesFrom{
			EdgesFrom: r,
		}
	case *apipb.DelNodesRequest:
		intercept.Request = &apipb.RequestIntercept_DelNodes{
			DelNodes: r,
		}
	case *apipb.DelEdgesRequest:
		intercept.Request = &apipb.RequestIntercept_DelEdges{
			DelEdges: r,
		}
	case *apipb.CreateNodesRequest:
		intercept.Request = &apipb.RequestIntercept_CreateNodes{
			CreateNodes: r,
		}
	case *apipb.CreateEdgesRequest:
		intercept.Request = &apipb.RequestIntercept_CreateEdges{
			CreateEdges: r,
		}
	case *apipb.JoinClusterRequest:
		intercept.Request = &apipb.RequestIntercept_JoinCluster{
			JoinCluster: r,
		}
	case *apipb.GetAuthRequest:
		intercept.Request = &apipb.RequestIntercept_GetAuth{
			GetAuth: r,
		}
	}
}
