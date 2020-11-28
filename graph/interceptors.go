package graph

import (
	"context"
	"encoding/json"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/helpers"
	"github.com/autom8ter/graphik/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	authCtxKey   = "x-graphik-auth-ctx"
	identityType = "identity"
	idClaim      = "sub"
	methodCtxKey = "x-grpc-full-method"
)

func (g *Graph) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
		if err != nil {
			return nil, err
		}
		payload, err := g.verifyJWT(token)
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
		ctx, identity, err := g.identityToContext(ctx, payload)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = g.methodToContext(ctx, info.FullMethod)
		if len(g.authorizers) > 0 {
			now := time.Now()
			request := &apipb.Request{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: timestamppb.New(now),
			}
			if val, ok := req.(proto.Message); ok {
				bits, _ := helpers.MarshalJSON(val)
				reqMap := map[string]interface{}{}
				if err := json.Unmarshal(bits, &reqMap); err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
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

func (g *Graph) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := grpc_auth.AuthFromMD(ss.Context(), "Bearer")
		if err != nil {
			return err
		}

		payload, err := g.verifyJWT(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, err.Error())
		}
		if val, ok := payload["exp"].(int64); ok && val < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		if val, ok := payload["exp"].(float64); ok && int64(val) < time.Now().Unix() {
			return status.Errorf(codes.Unauthenticated, "token expired")
		}
		ctx, identity, err := g.identityToContext(ss.Context(), payload)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		ctx = g.methodToContext(ss.Context(), info.FullMethod)
		if len(g.authorizers) > 0 {
			now := time.Now()
			request := &apipb.Request{
				Method:    info.FullMethod,
				Identity:  identity,
				Timestamp: timestamppb.New(now),
			}
			if val, ok := srv.(proto.Message); ok {
				bits, _ := helpers.MarshalJSON(val)
				reqMap := map[string]interface{}{}
				if err := json.Unmarshal(bits, &reqMap); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
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

func (a *Graph) identityToContext(ctx context.Context, payload map[string]interface{}) (context.Context, *apipb.Node, error) {
	path := &apipb.Path{
		Gtype: identityType,
		Gid:   payload[idClaim].(string),
	}
	var (
		node *apipb.Node
		err  error
	)
	if err = a.db.View(func(tx *bbolt.Tx) error {
		node, err = a.getNode(ctx, tx, path)
		return err
	}); err != nil && err != ErrNotFound {
		return ctx, nil, err
	}
	if node.GetPath() == nil {
		logger.Info("creating identity",
			zap.String("gtype", path.GetGtype()),
			zap.String("gid", path.GetGid()),
		)
		strct, err := structpb.NewStruct(payload)
		if err != nil {
			return nil, nil, err
		}
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
		return nil, nil, errors.New("empty node")
	}
	return context.WithValue(ctx, authCtxKey, node), node, nil
}

func (s *Graph) getIdentity(ctx context.Context) *apipb.Node {
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

func (r *Graph) getMethod(ctx context.Context) string {
	val, ok := ctx.Value(methodCtxKey).(string)
	if ok {
		return val
	}
	return ""
}

func (r *Graph) methodToContext(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, methodCtxKey, path)
}

func (g *Graph) verifyJWT(token string) (map[string]interface{}, error) {
	message, err := jws.ParseString(token)
	if err != nil {
		return nil, err
	}
	g.triggerMu.RLock()
	defer g.triggerMu.RUnlock()
	if g.jwksSet == nil {
		data := map[string]interface{}{}
		if err := json.Unmarshal(message.Payload(), &data); err != nil {
			return nil, err
		}
		return data, nil
	}
	if len(message.Signatures()) == 0 {
		return nil, fmt.Errorf("zero jws signatures")
	}
	kid, ok := message.Signatures()[0].ProtectedHeaders().Get("kid")
	if !ok {
		return nil, fmt.Errorf("jws kid not found")
	}
	algI, ok := message.Signatures()[0].ProtectedHeaders().Get("alg")
	if !ok {
		return nil, fmt.Errorf("jw alg not found")
	}
	alg, ok := algI.(jwa.SignatureAlgorithm)
	if !ok {
		return nil, fmt.Errorf("alg type cast error")
	}
	keys := g.jwksSet.LookupKeyID(kid.(string))
	if len(keys) == 0 {
		return nil, errors.Errorf("failed to lookup kid: %s - zero keys", kid.(string))
	}
	var key interface{}
	if err := keys[0].Raw(&key); err != nil {
		return nil, err
	}
	payload, err := jws.Verify([]byte(token), alg, key)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	return data, nil
}
