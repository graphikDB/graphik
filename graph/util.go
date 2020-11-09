package graph

import (
	"crypto/rand"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/mapstructure"
	"io"
	"strconv"
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

func FormPath(xtype, xid string) string {
	return strings.Join([]string{xtype, xid}, "/")
}

func SplitPath(path string) (string, string) {
	split := strings.Split(path, "/")
	if len(split) == 1 {
		return split[0], ""
	}
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}

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

func parseInt(obj interface{}) int {
	switch obj.(type) {
	case int:
		return obj.(int)
	case int32:
		return int(obj.(int32))
	case int64:
		return int(obj.(int64))
	case float32:
		return int(obj.(float32))
	case float64:
		return int(obj.(float64))
	case string:
		val, _ := strconv.Atoi(obj.(string))
		return val
	default:
		return 0
	}
}

func parseFloat(obj interface{}) float64 {
	switch o := obj.(type) {
	case int:
		return float64(o)
	case int32:
		return float64(o)
	case int64:
		return float64(o)
	case float32:
		return float64(o)
	case float64:
		return o
	case string:
		val, _ := strconv.ParseFloat(o, 64)
		return val
	default:
		return 0
	}
}

func parseString(obj interface{}) string {
	switch o := obj.(type) {
	case string:
		return o
	default:
		return fmt.Sprint(obj)
	}
}

func parseBool(obj interface{}) bool {
	switch obj.(type) {
	case bool:
		return obj.(bool)
	case string:
		val, _ := strconv.ParseBool(obj.(string))
		return val
	default:
		return false
	}
}
