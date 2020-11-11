package service

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/runtime"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	runtime *runtime.Runtime
}

func NewConfig(runtime *runtime.Runtime) *Config {
	return &Config{runtime: runtime}
}

func (s *Config) implements() apipb.ConfigServiceServer {
	return s
}

func (s *Config) JoinCluster(ctx context.Context, request *apipb.RaftNode) (*empty.Empty, error) {
	if err := s.runtime.JoinNode(request.GetNodeId(), request.GetAddress()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *Config) GetAuth(ctx context.Context, empty *empty.Empty) (*apipb.AuthConfig, error) {
	return s.runtime.Auth().Raw(), nil
}

func (s *Config) SetAuth(ctx context.Context, request *apipb.AuthConfig) (*apipb.AuthConfig, error) {
	if err := s.runtime.Auth().Override(request); err != nil {
		return nil, err
	}
	return s.runtime.Auth().Raw(), nil
}

func (s *Config) Ping(ctx context.Context, r *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}
