package helpers

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
	"strings"
)

func MarshalJSON(msg proto.Message) ([]byte, error) {
	return protojson.Marshal(msg)
}

func UnmarshalJSON(bits []byte, msg proto.Message) error {
	return protojson.Unmarshal(bits, msg)
}

func EnvOr(key string, defaul string) string {
	if val := os.Getenv(key); val == "" {
		return defaul
	} else {
		return val
	}
}

func StringSliceEnvOr(key string, defaul []string) []string {
	if value := os.Getenv(key); value != "" {
		values := strings.Split(value, ",")
		if len(value) > 0 && values[0] != "" {
			return values
		}
	} else {
		if len(defaul) > 0 && defaul[0] != "" {
			return defaul
		}
	}
	return nil
}

func BoolEnvOr(key string, defaul bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "y", "t", "yes":
			return true
		default:
			return false
		}
	}
	return defaul
}

func Hash(val []byte) string {
	h := sha1.New()
	h.Write(val)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func Uint64ToBytes(i uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, i)
	return buf
}

func BytesToUint64(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}
	return binary.BigEndian.Uint64(data)
}

func ContainsString(this string, arr []string) bool {
	for _, element := range arr {
		if element == this {
			return true
		}
	}
	return false
}
