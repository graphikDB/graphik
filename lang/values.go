package lang

import (
	"fmt"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"reflect"
	"time"
)

type Values map[string]interface{}

func NewValues(v map[string]interface{}) Values {
	return Values(v)
}

func (v Values) init() {
	if v == nil {
		v = map[string]interface{}{}
	}
}

func (v Values) GetType() string {
	v.init()
	return parseString(v[TypeKey])
}

func (v Values) GetID() string {
	v.init()
	return parseString(v[IDKey])
}

func (v Values) GetCreatedAt() int64 {
	v.init()
	return int64(parseInt(v[CreatedAtKey]))
}

func (v Values) GetUpdatedAt() int64 {
	v.init()
	return int64(parseInt(v[UpdatedAtKey]))
}

func (v Values) SetID(_id string) {
	v.init()
	v[IDKey] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: _id},
	}
}

func (v Values) SetType(_type string) {
	v.init()
	v[TypeKey] = _type
}

func (v Values) SetCreatedAt(_createdAt time.Time) {
	v.init()
	v[CreatedAtKey] = _createdAt.UnixNano()
}

func (v Values) SetUpdatedAt(_updatedAt time.Time) {
	v.init()
	v[UpdatedAtKey] = _updatedAt.UnixNano()
}

func (v Values) PathString() string {
	v.init()
	if v.GetID() == "" {
		return v.GetType()
	}
	return fmt.Sprintf("%s/%s", v.GetType(), v.GetID())
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
func (v Values) Copy(values *structpb.Struct) Values {
	copied := map[string]interface{}{}
	if values == nil {
		return copied
	}
	v.Range(func(k string, val interface{}) bool {
		v.Set(k, val)
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
