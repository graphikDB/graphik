package runtime

import (
	"context"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

const methodCtxKey = "x-grpc-full-method"

func (r *Runtime) UnaryAuth() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := r.Config().VerifyJWT(token)
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
		ctx, identity, err := r.ToContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = context.WithValue(ctx, methodCtxKey, info.FullMethod)
		if len(r.authorizers) > 0 {
			msg, ok := req.(proto.Message)
			if !ok {
				return nil, status.Error(codes.InvalidArgument, "bad request")
			}
			anny, err := ptypes.MarshalAny(msg)
			if err != nil {
				return nil, err
			}
			now := time.Now()
			for _, plugin := range r.authorizers {
				resp, err := plugin.Authorize(ctx, &apipb.RequestIntercept{
					Method:    info.FullMethod,
					User:      identity,
					Request:   anny,
					Timestamp: now.UnixNano(),
				})
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

func (r *Runtime) StreamAuth() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := grpc_auth.AuthFromMD(ss.Context(), "Bearer")
		if err != nil {
			return err
		}
		payload, err := r.Config().VerifyJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}
		if val, ok := payload["exp"].(int64); ok && val < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		if val, ok := payload["exp"].(float64); ok && int64(val) < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		ctx, identity, err := r.ToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = context.WithValue(ctx, methodCtxKey, info.FullMethod)
		if len(r.authorizers) > 0 {
			msg, ok := srv.(proto.Message)
			if !ok {
				return status.Error(codes.InvalidArgument, "bad request")
			}
			anny, err := ptypes.MarshalAny(msg)
			if err != nil {
				return err
			}
			now := time.Now()
			for _, plugin := range r.authorizers {
				resp, err := plugin.Authorize(ctx, &apipb.RequestIntercept{
					Method:    info.FullMethod,
					User:      identity,
					Request:   anny,
					Timestamp: now.UnixNano(),
				})
				if err != nil {
					return err
				}
				if !resp.Value {
					return status.Error(codes.PermissionDenied, "request authorization = denied")
				}
			}
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}
