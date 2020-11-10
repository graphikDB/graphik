package graph

import (
	"encoding/gob"
	"fmt"
	"github.com/autom8ter/graphik/sortable"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
	"strings"
	"time"
)

func init() {
	gob.Register(&Values{})
	gob.Register(&ValueSet{})
}

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

func (v Values) GetPath() string {
	v.init()
	return v.GetString(PathKey)
}

func (v Values) GetID() string {
	v.init()
	split := strings.Split(v.GetPath(), "/")
	if len(split) == 2 {
		return split[1]
	}
	return ""
}

func (v Values) GetType() string {
	v.init()
	split := strings.Split(v.GetPath(), "/")
	if len(split) > 0 {
		return split[0]
	}
	return ""
}

func (v Values) GetCreatedAt() int64 {
	v.init()
	return int64(parseInt(v[CreatedAtKey]))
}

func (v Values) GetUpdatedAt() int64 {
	v.init()
	return int64(parseInt(v[UpdatedAtKey]))
}

func (v Values) SetPath(path string) {
	v.init()
	v[PathKey] = path
}

func (v Values) SetType(xtype string) {
	v.init()
	if v.GetID() == "" {
		v[PathKey] = xtype
	}
	v[PathKey] = fmt.Sprintf("%s/%s", xtype, v.GetID())
}

func (v Values) SetID(xid string) {
	v.init()
	v[PathKey] = fmt.Sprintf("%s/%s", v.GetType(), xid)
}

func (v Values) SetCreatedAt(_createdAt time.Time) {
	v.init()
	v[CreatedAtKey] = _createdAt.UnixNano()
}

func (v Values) SetUpdatedAt(_updatedAt time.Time) {
	v.init()
	v[UpdatedAtKey] = _updatedAt.UnixNano()
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

type ValueSet []Values

func (v ValueSet) Structs() []*structpb.Struct {
	var structs []*structpb.Struct
	for _, val := range v {
		structs = append(structs, val.Struct())
	}
	return structs
}

func (v ValueSet) FromStructs(structs []*structpb.Struct) {
	for _, strct := range structs {
		v = append(v, strct.AsMap())
	}
}

func (v ValueSet) Sort() {
	s := &sortable.Sortable{
		LenFunc: func() int {
			return len(v)
		},
		LessFunc: func(i, j int) bool {
			return v[i].GetUpdatedAt() < v[j].GetUpdatedAt()
		},
		SwapFunc: func(i, j int) {
			v[i], v[j] = v[j], v[i]
		},
	}
	s.Sort()
}
