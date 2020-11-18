package plugins

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
)

type TriggerFunc func(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error)

func NewTrigger(trigger func(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error)) apipb.TriggerServiceServer {
	return TriggerFunc(trigger)
}

func (t TriggerFunc) HandleTrigger(ctx context.Context, trigger *apipb.Trigger) (*apipb.StateChange, error) {
	return t(ctx, trigger)
}

type AuthorizerFunc func(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error)

func NewAuthorizer(auth func(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error)) apipb.AuthorizationServiceServer {
	return AuthorizerFunc(auth)
}

func (a AuthorizerFunc) Authorize(ctx context.Context, intercept *apipb.RequestIntercept) (*apipb.Decision, error) {
	return a(ctx, intercept)
}
