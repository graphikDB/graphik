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

func (s *Service) SearchNodes(ctx context.Context, request *apipb.Filter) (*apipb.Nodes, error) {
	nodes, err := s.runtime.Nodes(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	nodes.Sort()
	return nodes, nil
}

func (s *Service) SearchEdges(ctx context.Context, request *apipb.Filter) (*apipb.Edges, error) {
	edges, err := s.runtime.Edges(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	edges.Sort()
	return edges, nil
}

func (s *Service) CreateNodes(ctx context.Context, request *apipb.Nodes) (*apipb.Nodes, error) {
	nodes, err := s.runtime.CreateNodes(request)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return nodes, nil
}

func (s *Service) PatchNodes(ctx context.Context, request *apipb.Patches) (*apipb.Nodes, error) {
	nodes, err := s.runtime.PatchNodes(request)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return nodes, nil
}

func (s *Service) DelNodes(ctx context.Context, request *apipb.Paths) (*apipb.Counter, error) {
	counter, err := s.runtime.DelNodes(request)
	if err != nil {
		return nil, err
	}
	return counter, nil
}

func (s *Service) CreateEdges(ctx context.Context, request *apipb.Edges) (*apipb.Edges, error) {
	edges, err := s.runtime.CreateEdges(request)
	if err != nil {
		return nil, err
	}
	return edges, nil
}

func (s *Service) PatchEdges(ctx context.Context, request *apipb.Patches) (*apipb.Edges, error) {
	edges, err := s.runtime.PatchEdges(request)
	if err != nil {
		return nil, err
	}
	return edges, nil
}

func (s *Service) DelEdges(ctx context.Context, request *apipb.Paths) (*apipb.Counter, error) {
	counter, err := s.runtime.DelEdges(request)
	if err != nil {
		return nil, err
	}
	return counter, nil
}

func (s *Service) EdgesFrom(ctx context.Context, request *apipb.PathFilter) (*apipb.Edges, error) {
	edges, err := s.runtime.EdgesFrom(request.GetPath(), request.GetFilter())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return edges, nil
}

func (s *Service) EdgesTo(ctx context.Context, request *apipb.PathFilter) (*apipb.Edges, error) {
	edges, err := s.runtime.EdgesTo(request.GetPath(), request.GetFilter())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return edges, nil
}

func (s *Service) Ping(ctx context.Context, r *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message: "PONG",
	}, nil
}
