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

func (s *Service) JoinCluster(ctx context.Context, request *apipb.JoinClusterRequest) (*empty.Empty, error) {
	if err := s.runtime.JoinNode(request.GetRaftNode().GetNodeId(), request.GetRaftNode().GetAddress()); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &empty.Empty{}, nil
}

func (s *Service) GetAuth(ctx context.Context, empty *empty.Empty) (*apipb.GetAuthResponse, error) {
	return &apipb.GetAuthResponse{
		Auth: s.runtime.Auth().Raw(),
	}, nil
}

func (s *Service) SetAuth(ctx context.Context, request *apipb.SetAuthRequest) (*apipb.SetAuthResponse, error) {
	if err := s.runtime.Auth().Override(request.GetAuth()); err != nil {
		return nil, err
	}
	return &apipb.SetAuthResponse{
		Auth: s.runtime.Auth().Raw(),
	}, nil
}

func (s *Service) SearchNodes(ctx context.Context, request *apipb.SearchNodesRequest) (*apipb.SearchNodesResponse, error) {
	nodes, err := s.runtime.Nodes(request.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	nodes.Sort()
	return &apipb.SearchNodesResponse{
		Nodes: nodes,
	}, nil
}

func (s *Service) SearchEdges(ctx context.Context, request *apipb.SearchEdgesRequest) (*apipb.SearchEdgesResponse, error) {
	edges, err := s.runtime.Edges(request.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	edges.Sort()
	return &apipb.SearchEdgesResponse{
		Edges: edges,
	}, nil
}

func (s *Service) CreateNodes(ctx context.Context, request *apipb.CreateNodesRequest) (*apipb.CreateNodesResponse, error) {
	nodes, err := s.runtime.CreateNodes(request.Nodes)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return &apipb.CreateNodesResponse{
		Nodes: nodes,
	}, nil
}

func (s *Service) PatchNodes(ctx context.Context, request *apipb.PatchNodesRequest) (*apipb.PatchNodesResponse, error) {
	nodes, err := s.runtime.PatchNodes(request.Patches)
	if err != nil {
		return nil, err
	}
	nodes.Sort()
	return &apipb.PatchNodesResponse{
		Nodes: nodes,
	}, nil
}

func (s *Service) DelNodes(ctx context.Context, request *apipb.DelNodesRequest) (*apipb.DelNodesResponse, error) {
	counter, err := s.runtime.DelNodes(request.Paths)
	if err != nil {
		return nil, err
	}
	return &apipb.DelNodesResponse{
		Counter: counter,
	}, nil
}

func (s *Service) CreateEdges(ctx context.Context, request *apipb.CreateEdgesRequest) (*apipb.CreateEdgesResponse, error) {
	edges, err := s.runtime.CreateEdges(request.Edges)
	if err != nil {
		return nil, err
	}
	return &apipb.CreateEdgesResponse{
		Edges: edges,
	}, nil
}

func (s *Service) PatchEdges(ctx context.Context, request *apipb.PatchEdgesRequest) (*apipb.PatchEdgesResponse, error) {
	edges, err := s.runtime.PatchEdges(request.Patches)
	if err != nil {
		return nil, err
	}
	return &apipb.PatchEdgesResponse{
		Edges: edges,
	}, nil
}

func (s *Service) DelEdges(ctx context.Context, request *apipb.DelEdgesRequest) (*apipb.DelEdgesResponse, error) {
	counter, err := s.runtime.DelEdges(request.Paths)
	if err != nil {
		return nil, err
	}
	return &apipb.DelEdgesResponse{
		Counter: counter,
	}, nil
}

func (s *Service) EdgesFrom(ctx context.Context, request *apipb.EdgesFromRequest) (*apipb.EdgesFromResponse, error) {
	edges, err := s.runtime.EdgesFrom(request.GetFilter().GetPath(), request.GetFilter().GetFilter())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &apipb.EdgesFromResponse{
		Edges: edges,
	}, nil
}

func (s *Service) EdgesTo(ctx context.Context, request *apipb.EdgesToRequest) (*apipb.EdgesToResponse, error) {
	edges, err := s.runtime.EdgesTo(request.GetFilter().GetPath(), request.GetFilter().GetFilter())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &apipb.EdgesToResponse{
		Edges: edges,
	}, nil
}

func (s *Service) Ping(ctx context.Context, r *empty.Empty) (*apipb.Pong, error) {
	return &apipb.Pong{
		Message:              "PONG",
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}, nil
}
