package graph

import (
	"context"
	"encoding/json"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/helpers"
	"github.com/autom8ter/graphik/logger"
	"github.com/golang/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
	methodCtxKey = "x-grpc-full-method"
)

type intercept struct {
	Method    string
	Identity  map[string]interface{}
	Timestamp *timestamppb.Timestamp
	Request   map[string]interface{}
	Response  map[string]interface{}
}

func (r *intercept) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"method":    r.Method,
		"identity":  r.Identity,
		"request":   r.Request,
		"timestamp": r.Timestamp,
	}
}

func (g *GraphStore) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := g.VerifyJWT(token)
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
		ctx, identity, err := g.NodeToContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = g.MethodToContext(ctx, info.FullMethod)
		if len(g.authorizers) > 0 {
			now := time.Now()
			request := &apipb.Request{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: timestamppb.New(now),
			}
			if val, ok := req.(proto.Message); ok {
				str := helpers.JSONString(val)
				reqMap := map[string]interface{}{}
				json.NewDecoder(strings.NewReader(str)).Decode(&reqMap)
				request.Request = apipb.NewStruct(reqMap)
			}
			result, err := g.vm.Auth().Eval(g.authorizers, request)
			if err != nil {
				return nil, err
			}
			if !result {
				return nil, status.Error(codes.PermissionDenied, "request authorization = denied")
			}
		}
		return handler(ctx, req)
	}
}

func (g *GraphStore) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := grpc_auth.AuthFromMD(ss.Context(), "Bearer")
		if err != nil {
			return err
		}

		payload, err := g.VerifyJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}
		if val, ok := payload["exp"].(int64); ok && val < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		if val, ok := payload["exp"].(float64); ok && int64(val) < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		ctx, identity, err := g.NodeToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = g.MethodToContext(ss.Context(), info.FullMethod)
		if len(g.authorizers) > 0 {
			now := time.Now()
			request := &apipb.Request{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: timestamppb.New(now),
			}
			if val, ok := srv.(proto.Message); ok {
				str := helpers.JSONString(val)
				reqMap := map[string]interface{}{}
				json.NewDecoder(strings.NewReader(str)).Decode(&reqMap)
				request.Request = apipb.NewStruct(reqMap)
			}
			result, err := g.vm.Auth().Eval(g.authorizers, request)
			if err != nil {
				return err
			}
			if !result {
				return status.Error(codes.PermissionDenied, "request authorization = denied")
			}
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}

func (a *GraphStore) NodeToContext(ctx context.Context, payload map[string]interface{}) (context.Context, *apipb.Node, error) {
	path := &apipb.Path{
		Gtype: identityType,
		Gid:   payload[idClaim].(string),
	}
	var node = &apipb.Node{}
	if err := a.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbNodes).Bucket([]byte(path.Gtype))
		if bucket == nil {
			return ErrNotFound
		}
		bits := bucket.Get([]byte(path.Gid))
		if err := proto.Unmarshal(bits, node); err != nil {
			return err
		}
		return nil
	}); err != nil && err != ErrNotFound {
		return ctx, nil, err
	}

	if node.GetPath() == nil {
		logger.Info("creating identity",
			zap.String("gtype", path.GetGtype()),
			zap.String("gid", path.GetGid()),
		)
		strct, _ := structpb.NewStruct(payload)
		nodeP, err := a.createIdentity(ctx, &apipb.NodeConstructor{
			Path:       path,
			Attributes: strct,
		})
		if err != nil {
			return nil, nil, err
		}
		node = nodeP
	}
	if node.GetPath() == nil {
		panic("empty node")
	}
	return context.WithValue(ctx, authCtxKey, node), node, nil
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
