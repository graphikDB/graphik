package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
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
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers:  r,
		Directives: generated.DirectiveRoot{},
		Complexity: generated.ComplexityRoot{},
	}))
}
