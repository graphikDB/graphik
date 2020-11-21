package gql

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/gql/generated"
	"github.com/autom8ter/machine"
	"github.com/rs/cors"
	"google.golang.org/grpc/metadata"
	"net/http"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client  *graphik.Client
	cors    *cors.Cors
	machine *machine.Machine
}

func NewResolver(ctx context.Context, client *graphik.Client, cors *cors.Cors) *Resolver {
	return &Resolver{client: client, cors: cors, machine: machine.New(ctx)}
}

func (r *Resolver) QueryHandler() http.Handler {
	return r.cors.Handler(r.authMiddleware(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	}))))
}

func (r *Resolver) authMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		for k, arr := range req.Header {
			if len(arr) > 0 {
				ctx = metadata.AppendToOutgoingContext(ctx, k, arr[0])
			}
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	}
}
