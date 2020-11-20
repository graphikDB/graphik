package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/gql/generated"
	"golang.org/x/oauth2"
	"net/http"
)

// This file will not be regenerated automatically.
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client *graphik.Client
	config *oauth2.Config
}

func NewResolver(client *graphik.Client) *Resolver {
	return &Resolver{client: client}
}

func (r *Resolver) QueryHandler() http.Handler {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	}))
}

func (r *Resolver) Playground(endpoint string) http.Handler {
	return playground.Handler("Graphik Playground", endpoint)
}

const usrCtx = "user-context"

func (r *Resolver) AuthMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handler.ServeHTTP(w, req)
	}
}

