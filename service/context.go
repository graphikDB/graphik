package service

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/ptypes/empty"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
	methodCtxKey = "x-grpc-full-method"
)

func (r *GraphStore) UnaryAuth() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := r.auth.VerifyJWT(token)
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
		ctx, identity, err := r.NodeToContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = r.MethodToContext(ctx, info.FullMethod)
		if len(r.authorizers) > 0 {
			now := time.Now()
			intercept := &apipb.RequestIntercept{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: now.UnixNano(),
			}
			switch r := req.(type) {
			case *apipb.NodeConstructor:
				intercept.Request = &apipb.RequestIntercept_NodeConstructor{NodeConstructor: r}
			case *apipb.NodeConstructors:
				intercept.Request = &apipb.RequestIntercept_NodeConstructors{NodeConstructors: r}
			case *apipb.Paths:
				intercept.Request = &apipb.RequestIntercept_Paths{Paths: r}
			case *apipb.Path:
				intercept.Request = &apipb.RequestIntercept_Path{Path: r}
			case *apipb.Patches:
				intercept.Request = &apipb.RequestIntercept_Patches{Patches: r}
			case *apipb.Patch:
				intercept.Request = &apipb.RequestIntercept_Patch{Patch: r}
			case *apipb.Filter:
				intercept.Request = &apipb.RequestIntercept_Filter{Filter: r}
			case *apipb.ChannelFilter:
				intercept.Request = &apipb.RequestIntercept_ChannelFilter{ChannelFilter: r}
			case *apipb.RaftNode:
				intercept.Request = &apipb.RequestIntercept_RaftNode{RaftNode: r}
			case *apipb.SubGraphFilter:
				intercept.Request = &apipb.RequestIntercept_SubgraphFilter{SubgraphFilter: r}
			case *empty.Empty:
				intercept.Request = &apipb.RequestIntercept_Empty{Empty: r}
			}
			for _, plugin := range r.authorizers {
				resp, err := plugin.Authorize(ctx, intercept)
				if err != nil {
					return nil, err
				}
				if !resp.Value {
					return nil, status.Error(codes.PermissionDenied, "request authorization = denied")
				}
			}
		}
		return handler(ctx, req)
	}
}

func (r *GraphStore) StreamAuth() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := grpc_auth.AuthFromMD(ss.Context(), "Bearer")
		if err != nil {
			return err
		}
		payload, err := r.auth.VerifyJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}
		if val, ok := payload["exp"].(int64); ok && val < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		if val, ok := payload["exp"].(float64); ok && int64(val) < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		ctx, identity, err := r.NodeToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = r.MethodToContext(ss.Context(), info.FullMethod)
		if len(r.authorizers) > 0 {
			now := time.Now()
			intercept := &apipb.RequestIntercept{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: now.UnixNano(),
			}
			switch r := srv.(type) {
			case *apipb.NodeConstructor:
				intercept.Request = &apipb.RequestIntercept_NodeConstructor{NodeConstructor: r}
			case *apipb.NodeConstructors:
				intercept.Request = &apipb.RequestIntercept_NodeConstructors{NodeConstructors: r}
			case *apipb.Paths:
				intercept.Request = &apipb.RequestIntercept_Paths{Paths: r}
			case *apipb.Path:
				intercept.Request = &apipb.RequestIntercept_Path{Path: r}
			case *apipb.Patches:
				intercept.Request = &apipb.RequestIntercept_Patches{Patches: r}
			case *apipb.Patch:
				intercept.Request = &apipb.RequestIntercept_Patch{Patch: r}
			case *apipb.Filter:
				intercept.Request = &apipb.RequestIntercept_Filter{Filter: r}
			case *apipb.ChannelFilter:
				intercept.Request = &apipb.RequestIntercept_ChannelFilter{ChannelFilter: r}
			case *apipb.RaftNode:
				intercept.Request = &apipb.RequestIntercept_RaftNode{RaftNode: r}
			case *apipb.SubGraphFilter:
				intercept.Request = &apipb.RequestIntercept_SubgraphFilter{SubgraphFilter: r}
			case *empty.Empty:
				intercept.Request = &apipb.RequestIntercept_Empty{Empty: r}
			}
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}

func (a *GraphStore) NodeToContext(ctx context.Context, payload map[string]interface{}) (context.Context, *apipb.Node, error) {
	var err error
	n, err := a.GetNode(ctx, &apipb.Path{
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

func (s *GraphStore) NodeContext(ctx context.Context) *apipb.Node {
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

func (r *GraphStore) MethodContext(ctx context.Context) string {
	val, ok := ctx.Value(methodCtxKey).(string)
	if ok {
		return val
	}
	return ""
}

func (r *GraphStore) MethodToContext(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, methodCtxKey, path)
}
