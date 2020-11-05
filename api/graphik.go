package graphikpb

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/autom8ter/graphik/lib/helpers"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"io"
)

var marshaler = jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: false,
	Indent:       "",
	OrigName:     false,
	AnyResolver:  nil,
}

func MarshalNode(val *structpb.Struct) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if val != nil {
			marshaler.Marshal(w, val)
		}
	})
}

func UnmarshalNode(v interface{}) (*structpb.Struct, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		return helpers.ToStruct(v), nil
	default:
		return nil, fmt.Errorf("%T is not a map[string]interface{}", v)
	}
}