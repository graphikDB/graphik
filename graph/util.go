package graph

import (
	"crypto/rand"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/mapstructure"
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
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
	case map[string]interface{}:
		return o
	default:
		values := map[string]interface{}{}
		mapstructure.WeakDecode(o, &values)
		return values
	}
}
