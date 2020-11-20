package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/gql/generated"
	"google.golang.org/grpc/metadata"
	"net/http"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client *graphik.Client
}

func NewResolver(client *graphik.Client) *Resolver {
	return &Resolver{client: client}
}

func (r *Resolver) QueryHandler() http.Handler {
	return r.authMiddleware(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	})))
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