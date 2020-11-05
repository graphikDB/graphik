package admin

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lib/runtime"
)

type Service struct {
	runtime *runtime.Runtime
}

func NewService(runtime *runtime.Runtime) *Service {
	return &Service{runtime: runtime}
}

func (s Service) RaftJoin(ctx context.Context, request *apipb.RaftJoinRequest) (*apipb.RaftJoinResponse, error) {
	panic("implement me")
}

func (s Service) GetJWKS(ctx context.Context, request *apipb.GetJWKSRequest) (*apipb.GetJWKSResponse, error) {
	panic("implement me")
}

func (s Service) UpdateJWKS(ctx context.Context, request *apipb.UpdateJWKSRequest) (*apipb.UpdateJWKSResponse, error) {
	panic("implement me")
}
