package graph

import (
	"crypto/rand"
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
)

var (
	marshaller = &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "",
		OrigName:     false,
		AnyResolver:  nil,
	}
	unmarshaller = &jsonpb.Unmarshaler{}
)

func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func JSONEncode(w io.Writer, msg proto.Message) error {
	return marshaller.Marshal(w, msg)
}

func JSONDecode(r io.Reader, msg proto.Message) error {
	return unmarshaller.Unmarshal(r, msg)
}

func ToMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case map[string]interface{}:
		return o
	case *apipb.Node:
		return map[string]interface{}{
			"path": map[string]interface{}{
				"gid":   o.GetPath().GetGid(),
				"gtype": o.GetPath().GetGtype(),
			},
			"attributes": o.GetAttributes(),
			"created_at": o.GetCreatedAt(),
			"updated_at": o.GetUpdatedAt(),
		}
	case *apipb.Edge:
		return map[string]interface{}{
			"path": map[string]interface{}{
				"gid":   o.GetPath().GetGid(),
				"gtype": o.GetPath().GetGtype(),
			},
			"attributes": o.GetAttributes(),
			"cascade":    o.GetCascade().String(),
			"from": map[string]interface{}{
				"gid":  o.GetFrom().GetGid(),
				"type": o.GetFrom().GetGtype(),
			},
			"to": map[string]interface{}{
				"gid":   o.GetFrom().GetGid(),
				"gtype": o.GetFrom().GetGtype(),
			},
			"created_at": o.GetCreatedAt(),
			"updated_at": o.GetUpdatedAt(),
		}
	case *structpb.Struct:
		return o.AsMap()
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
	default:
		values := map[string]interface{}{}
		mapstructure.WeakDecode(o, &values)
		return values
	}
}
