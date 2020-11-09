package private

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/runtime"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	runtime *runtime.Runtime
}

func NewService(runtime *runtime.Runtime) *Service {
	return &Service{runtime: runtime}
}

func (s *Service) implements() apipb.PrivateServiceServer {
	return s
}

func (s *Service) JoinCluster(ctx context.Context, request *apipb.RaftNode) (*empty.Empty, error) {
	if err := s.runtime.JoinNode(request.GetNodeId(), request.GetAddress()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *Service) GetAuth(ctx context.Context, empty *empty.Empty) (*apipb.AuthConfig, error) {
	return s.runtime.Auth().Raw(), nil
}

func (s *Service) SetAuth(ctx context.Context, request *apipb.AuthConfig) (*apipb.AuthConfig, error) {
	if err := s.runtime.Auth().Override(request); err != nil {
		return nil, err
	}
	return s.runtime.Auth().Raw(), nil
}

func (s *Service) Ping(ctx context.Context, r *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}
