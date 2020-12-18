package helpers

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
	"strconv"
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

func IntEnvOr(key string, defaul int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
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

func ContainsString(this string, arr []string) bool {
	for _, element := range arr {
		if element == this {
			return true
		}
	}
	return false
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func JSONString(obj interface{}) string {
	bits, _ := json.MarshalIndent(obj, "", "    ")
	return string(bits)
}
