package scalars

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/helpers"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"strings"
	"time"
)

func MarshalStructScalar(n *structpb.Struct) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(helpers.JSONString(n)))
	})
}

func UnmarshalStructScalar(v interface{}) (*structpb.Struct, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		return structpb.NewStruct(v)
	case string:
		val := structpb.Struct{}
		return &val, jsonpb.Unmarshal(strings.NewReader(v), &val)
	case *structpb.Struct:
		return v, nil
	case structpb.Struct:
		return &v, nil
	default:
		return nil, fmt.Errorf("%T is not a struct", v)
	}
}

func MarshalEmptyScalar(n *empty.Empty) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(helpers.JSONString(n)))
	})
}

func UnmarshalEmptyScalar(v interface{}) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func UnmarshalCascadeScalar(v interface{}) (apipb.Cascade, error) {
	switch v := v.(type) {
	case string:
		return apipb.Cascade(apipb.Cascade_value[v]), nil
	case int:
		return apipb.Cascade(v), nil
	case int32:
		return apipb.Cascade(v), nil
	default:
		return apipb.Cascade_CASCADE_NONE, fmt.Errorf("%T is not a Cascade", v)
	}
}

func MarshalCascadeScalar(c apipb.Cascade) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(c.String()))
	})
}

func MarshalTimestampScalar(t *timestamp.Timestamp) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(t.AsTime().String()))
	})
}

func UnmarshalTimestampScalar(v interface{}) (*timestamp.Timestamp, error) {
	switch v := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return timestamppb.New(t), nil
	case int:
		return timestamppb.New(time.Unix(int64(v), 0)), nil
	case int32:
		return timestamppb.New(time.Unix(int64(v), 0)), nil
	case int64:
		return timestamppb.New(time.Unix(int64(v), 0)), nil
	default:
		return nil, fmt.Errorf("%T is not a Cascade", v)
	}
}
