package helpers

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"io"
	"os"
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

func EnvOr(key string, defaul string) string {
	if val := os.Getenv(key); val == "" {
		return defaul
	} else {
		return val
	}
}
