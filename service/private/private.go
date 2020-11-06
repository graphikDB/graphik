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
	var toReturn []*apipb.JWKSSource
	sources := s.runtime.JWKS().List()
	for uri, issuer := range sources {
		toReturn = append(toReturn, &apipb.JWKSSource{
			Uri:    uri,
			Issuer: issuer,
		})
	}
	return &apipb.GetJWKSResponse{
		Sources: toReturn,
	}, nil
}

func (s *Service) UpdateJWKS(ctx context.Context, request *apipb.UpdateJWKSRequest) (*apipb.UpdateJWKSResponse, error) {
	sourceMap := map[string]string{}
	for _, source := range request.Sources {
		sourceMap[source.Uri] = source.Issuer
	}
	if err := s.runtime.JWKS().Override(sourceMap); err != nil {
		return nil, err
	}
	var toReturn []*apipb.JWKSSource
	sources := s.runtime.JWKS().List()
	for uri, issuer := range sources {
		toReturn = append(toReturn, &apipb.JWKSSource{
			Uri:    uri,
			Issuer: issuer,
		})
	}
	return &apipb.UpdateJWKSResponse{
		Sources: toReturn,
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
	edges, err := s.runtime.EdgesFrom(request.Path, request.Filter)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &apipb.EdgesFromResponse{
		Edges: edges,
	}, nil
}

func (s *Service) EdgesTo(ctx context.Context, request *apipb.EdgesToRequest) (*apipb.EdgesToResponse, error) {
	edges, err := s.runtime.EdgesTo(request.Path, request.Filter)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &apipb.EdgesToResponse{
		Edges: edges,
	}, nil
}
