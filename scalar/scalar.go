package scalar

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/autom8ter/dagger"
)

// Lets redefine the base ID type to use an id from an external library
func MarshalNode(node *dagger.Node) graphql.Marshaler {
	return graphql.MarshalMap(node.Raw())
}

// And the same for the unmarshaler
func UnmarshalNode(v interface{}) (*dagger.Node, error) {
	data, err := graphql.UnmarshalMap(v)
	if err != nil {
		return nil, err
	}
	return dagger.NewNode(data), err
}