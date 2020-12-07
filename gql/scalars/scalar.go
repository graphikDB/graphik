package scalars

import (
	"github.com/99designs/gqlgen/graphql"
	apipb "github.com/autom8ter/graphik/gen/go"
	"github.com/golang/protobuf/ptypes/empty"
	"io"
)

func MarshalEmptyScalar(n *empty.Empty) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte("{}"))
	})
}

func UnmarshalEmptyScalar(v interface{}) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func MarshalDirectionScalar(n apipb.Direction) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(n.String()))
	})
}

func UnmarshalDirectionScalar(v interface{}) (apipb.Direction, error) {
	switch val := v.(type) {
	case string:
		return apipb.Direction(apipb.Direction_value[val]), nil
	case int32:
		return apipb.Direction(val), nil
	case apipb.Direction:
		return val, nil
	default:
		return apipb.Direction_None, nil
	}
}
