package private

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/runtime"
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

func (s *Service) JoinCluster(ctx context.Context, request *apipb.JoinClusterRequest) (*apipb.JoinClusterResponse, error) {
	return &apipb.JoinClusterResponse{}, s.runtime.JoinNode(request.NodeId, request.Address)
}

func (s *Service) GetJWKS(ctx context.Context, request *apipb.GetJWKSRequest) (*apipb.GetJWKSResponse, error) {
	return &apipb.GetJWKSResponse{
		Jwks:                 s.runtime.JWKS().List(),
	}, nil
}

func (s *Service) UpdateJWKS(ctx context.Context, request *apipb.UpdateJWKSRequest) (*apipb.UpdateJWKSResponse, error) {
	if err := s.runtime.JWKS().Override(request.Jwks); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &apipb.UpdateJWKSResponse{
		Jwks:                 s.runtime.JWKS().List(),
	}, nil
}

func (s *Service) SearchNodes(ctx context.Context, request *apipb.SearchNodesRequest) (*apipb.SearchNodesResponse, error) {
	nodes, err := s.runtime.Nodes(request.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &apipb.SearchNodesResponse{
		Nodes:                nodes,
	}, nil
}

func (s *Service) SearchEdges(ctx context.Context, request *apipb.SearchEdgesRequest) (*apipb.SearchEdgesResponse, error) {
	edges, err := s.runtime.Edges(request.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &apipb.SearchEdgesResponse{
		Edges:                edges,
	}, nil
}

func (s *Service) CreateNodes(ctx context.Context, request *apipb.CreateNodesRequest) (*apipb.CreateNodesResponse, error) {
	var nodes []*apipb.Node
	for _, n  := range request.Nodes {
		node, err := s.runtime.CreateNode(&apipb.Node{
			Path:                 n.Path,
			Attributes:           n.Attributes,
		})
		if err != nil {
			 return nil, err
		}
		nodes = append(nodes, node)
	}
	return &apipb.CreateNodesResponse{
		Nodes:                nodes,
	}, nil
}

func (s *Service) PatchNodes(ctx context.Context, request *apipb.PatchNodesRequest) (*apipb.PatchNodesResponse, error) {
	var nodes []*apipb.Node
	for _, n  := range request.Patches {
		node, err := s.runtime.PatchNode(&apipb.Patch{
			Path:                 n.Path,
			Patch:           n.Patch,
		})
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return &apipb.PatchNodesResponse{
		Node:                nodes,
	}, nil
}

func (s *Service) DelNodes(ctx context.Context, request *apipb.DelNodesRequest) (*apipb.DelNodesResponse, error) {
	panic("implement me")
}

func (s *Service) CreateEdges(ctx context.Context, request *apipb.CreateEdgesRequest) (*apipb.CreateEdgesResponse, error) {
	panic("implement me")
}

func (s *Service) PatchEdges(ctx context.Context, request *apipb.PatchEdgesRequest) (*apipb.PatchEdgesResponse, error) {
	panic("implement me")
}

func (s *Service) DelEdges(ctx context.Context, request *apipb.DelEdgesRequest) (*apipb.DelEdgesResponse, error) {
	panic("implement me")
}
