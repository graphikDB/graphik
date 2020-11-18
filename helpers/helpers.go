package helpers

import (
	"bytes"
	"encoding/json"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/mitchellh/mapstructure"
	"io"
	"strings"
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

func JSONEncode(w io.Writer, msg proto.Message) error {
	return marshaller.Marshal(w, msg)
}

func JSONDecode(r io.Reader, msg proto.Message) error {
	return unmarshaller.Unmarshal(r, msg)
}

func JSONString(msg proto.Message) string {
	buf := bytes.NewBuffer(nil)
	JSONEncode(buf, msg)
	return buf.String()
}

func FromJSONString(str string, msg proto.Message) error {
	return JSONDecode(strings.NewReader(str), msg)
}

func ToMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case nil:
		return map[string]interface{}{}
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
	case *empty.Empty:
		return map[string]interface{}{}
	case *structpb.Struct:
		return o.AsMap()
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
	case proto.Message:
		buf := bytes.NewBuffer(nil)
		var data = map[string]interface{}{}
		JSONEncode(buf, o)
		json.Unmarshal(buf.Bytes(), &data)
		return data
	default:
		values := map[string]interface{}{}
		mapstructure.WeakDecode(o, &values)
		return values
	}
}
