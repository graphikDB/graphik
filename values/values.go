package values

import (
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
	"strconv"
)

type Values map[string]interface{}

func NewValues(v map[string]interface{}) Values {
	return v
}

func (v Values) init() {
	if v == nil {
		v = map[string]interface{}{}
	}
}

func (v Values) FromStruct(strct *structpb.Struct) {
	for k, val := range strct.AsMap() {
		v[k] = val
	}
}

// Set set an entry in the Node
func (v Values) Set(k string, val interface{}) {
	v.init()
	v[k] = val
}

// SetAll set all entries in the Node
func (v Values) SetAll(data map[string]interface{}) {
	v.init()
	if data == nil {
		return
	}
	for k, val := range data {
		v.Set(k, val)
	}
}

// Get gets an entry from the Node by key
func (v Values) Get(key string) interface{} {
	v.init()
	return v[key]
}

// Exists returns true if the key exists in the Node
func (v Values) Exists(key string) bool {
	v.init()
	if val, ok := v[key]; ok && val != nil {
		return true
	}
	return false
}

// GetString gets an entry from the Node by key
func (v Values) GetString(key string) string {
	v.init()
	if !v.Exists(key) {
		return ""
	}
	return parseString(v[key])
}

func (v Values) GetBool(key string) bool {
	if !v.Exists(key) {
		return false
	}
	return parseBool(v[key])
}

func (v Values) GetInt(key string) int {
	if !v.Exists(key) {
		return 0
	}
	return parseInt(v[key])
}

func (v Values) GetFloat(key string) float64 {
	if !v.Exists(key) {
		return 0
	}
	return parseFloat(v[key])
}

// Del deletes the entry from the Node by key
func (v Values) Del(key string) {
	v.init()
	delete(v, key)
}

// Range iterates over the Node with the function. If the function returns false, the iteration exits.
func (v Values) Range(iterator func(key string, val interface{}) bool) {
	v.init()
	for k, val := range v {
		if !iterator(k, val) {
			break
		}
	}
}

// Filter returns a Node of the node that return true from the filter function
func (v Values) Filter(filter func(key string, v interface{}) bool) Values {
	v.init()
	data := map[string]interface{}{}
	v.Range(func(key string, val interface{}) bool {
		if filter(key, val) {
			data[key] = val
		}
		return true
	})
	return data
}

// Copy creates a replica of the Node
func (v Values) Copy() Values {
	copied := map[string]interface{}{}
	if v == nil {
		return copied
	}
	v.Range(func(k string, val interface{}) bool {
		copied[k] = val
		return true
	})
	return copied
}

func (v Values) Equals(other Values) bool {
	return reflect.DeepEqual(v, other)
}

func (v Values) GetNested(key string) (Values, bool) {
	if val, ok := v[key]; ok && val != nil {
		if values, ok := val.(map[string]interface{}); ok {
			return values, true
		}
		if values, ok := val.(Values); ok {
			return values, true
		}
	}
	return nil, false
}

func (v Values) IsNested(key string) bool {
	_, ok := v.GetNested(key)
	return ok
}

func (v Values) SetNested(key string, nested Values) {
	v.Set(key, nested)
}

func (v Values) Struct() *structpb.Struct {
	s, err := structpb.NewStruct(v)
	if err != nil {
		panic(err)
	}
	return s
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
