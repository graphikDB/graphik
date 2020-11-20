package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/gql/generated"
	"net/http"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client *graphik.Client
}

func NewResolver(client *graphik.Client) *Resolver {
	return &Resolver{client: client}
}

func (r *Resolver) Handler() http.Handler {
	return r.authMiddleware(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	})))
}

func (r *Resolver) Playground() http.Handler {
	return r.authMiddleware(playground.Handler("GraphQL playground", "/query"))
}

func (r *Resolver) authMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
