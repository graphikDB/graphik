package scalars

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/autom8ter/graphik/helpers"
	"github.com/golang/protobuf/ptypes/empty"
	"io"
)

func MarshalEmptyScalar(n *empty.Empty) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		bits, _ := helpers.MarshalJSON(n)
		w.Write(bits)
	})
}

func UnmarshalEmptyScalar(v interface{}) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
