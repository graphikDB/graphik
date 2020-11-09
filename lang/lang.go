package lang

import (
	"crypto/rand"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"io"
	"reflect"
	"strings"
)

const (
	IDKey        = "_id"
	TypeKey      = "_type"
	CreatedAtKey = "_createdAt"
	UpdatedAtKey = "_updatedAt"
	FromKey = "_from"
	ToKey = "_to"
	CascadeKey = "_cascade"
	MutualKey = "_mutual"
)

type Values struct {
	*structpb.Struct
}

func NewValues(values *structpb.Struct) *Values {
	return &Values{values}
}

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

// ToStruct converts a map[string]interface{} to a ptypes.Struct
func ToStruct(v map[string]interface{}) *Values {
	size := len(v)
	if size == 0 {
		return nil
	}
	fields := make(map[string]*structpb.Value, size)
	for k, v := range v {
		fields[k] = ToValue(v)
	}
	return &Values{
		Struct: &structpb.Struct{
			Fields: fields,
		}}
}

func (v *Values) Map() map[string]interface{} {
	values := map[string]interface{}{}
	for k, field := range v.Fields {
		values[k] = FromValue(field)
	}
	return values
}

func FromValue(field *structpb.Value) interface{} {
	switch field.GetKind().(type) {
	case *structpb.Value_BoolValue:
		return field.GetBoolValue()
	case *structpb.Value_NumberValue:
		return field.GetNumberValue()
	case *structpb.Value_NullValue:
		return field.GetNullValue()
	case *structpb.Value_StringValue:
		return field.GetStringValue()
	case *structpb.Value_StructValue:
		vals := &Values{field.GetStructValue()}
		return vals.Map()
	case *structpb.Value_ListValue:
		var values []interface{}
		for _, v := range field.GetListValue().GetValues() {
			values = append(values, FromValue(v))
		}
		return values
	default:
		return nil
	}
}

// ToValue converts an interface{} to a ptypes.Value
func ToValue(v interface{}) *structpb.Value {
	switch v := v.(type) {
	case nil:
		return &structpb.Value{
			Kind: &structpb.Value_NullValue{
				NullValue: structpb.NullValue_NULL_VALUE,
			},
		}
	case bool:
		return &structpb.Value{
			Kind: &structpb.Value_BoolValue{
				BoolValue: v,
			},
		}
	case int:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int8:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case int64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint8:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case uint64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float32:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v),
			},
		}
	case float64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: v,
			},
		}
	case string:
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v,
			},
		}
	case error:
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v.Error(),
			},
		}
	case *structpb.Struct:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: v}}
	case structpb.Struct:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &v}}
	case *Values:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: v.Struct}}
	case Values:
		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: v.Struct}}
	default:
		// Fallback to reflection for other types
		return toValue(reflect.ValueOf(v))
	}
}

func toValue(v reflect.Value) *structpb.Value {
	switch v.Kind() {
	case reflect.Bool:
		return &structpb.Value{
			Kind: &structpb.Value_BoolValue{
				BoolValue: v.Bool(),
			},
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v.Int()),
			},
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(v.Uint()),
			},
		}
	case reflect.Float32, reflect.Float64:
		return &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: v.Float(),
			},
		}
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return toValue(reflect.Indirect(v))
	case reflect.Array, reflect.Slice:
		size := v.Len()
		if size == 0 {
			return nil
		}
		values := make([]*structpb.Value, size)
		for i := 0; i < size; i++ {
			values[i] = toValue(v.Index(i))
		}
		return &structpb.Value{
			Kind: &structpb.Value_ListValue{
				ListValue: &structpb.ListValue{
					Values: values,
				},
			},
		}
	case reflect.Struct:
		t := v.Type()
		size := v.NumField()
		if size == 0 {
			return nil
		}
		fields := make(map[string]*structpb.Value, size)
		for i := 0; i < size; i++ {
			name := t.Field(i).Name
			// Better way?
			if len(name) > 0 && 'A' <= name[0] && name[0] <= 'Z' {
				fields[name] = toValue(v.Field(i))
			}
		}
		if len(fields) == 0 {
			return nil
		}
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Map:
		keys := v.MapKeys()
		if len(keys) == 0 {
			return nil
		}
		fields := make(map[string]*structpb.Value, len(keys))
		for _, k := range keys {
			if k.Kind() == reflect.String {
				fields[k.String()] = toValue(v.MapIndex(k))
			}
		}
		if len(fields) == 0 {
			return nil
		}
		return &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: &structpb.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Interface:
		return ToValue(v.Interface())
	default:
		// Last resort
		return &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: fmt.Sprint(v),
			},
		}
	}
}

func (v *Values) init() {
	if v == nil {
		v = &Values{&structpb.Struct{
			Fields: map[string]*structpb.Value{},
		}}
	}
	if v.Fields == nil {
		v.Fields = map[string]*structpb.Value{}
	}
}

func (v *Values) GetType() string {
	v.init()
	return v.GetFields()[TypeKey].GetStringValue()
}

func (v *Values) GetID() string {
	v.init()
	return v.GetFields()[IDKey].GetStringValue()
}

func (v *Values) GetCreatedAt() int64 {
	v.init()
	return int64(v.GetFields()[CreatedAtKey].GetNumberValue())
}

func (v *Values) GetUpdatedAt() int64 {
	v.init()
	return int64(v.GetFields()[UpdatedAtKey].GetNumberValue())
}

func (v *Values) SetID(_id string) {
	v.init()
	v.GetFields()[IDKey] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: _id},
	}
}

func (v *Values) SetType(_type string) {
	v.init()
	v.GetFields()[TypeKey] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: _type},
	}
}

func (v *Values) SetCreatedAt(_createdAt *timestamp.Timestamp) {
	v.init()
	v.GetFields()[CreatedAtKey] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(_createdAt.Seconds)},
	}
}

func (v *Values) SetUpdatedAt(_updatedAt *timestamp.Timestamp) {
	v.init()
	v.GetFields()[UpdatedAtKey] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(_updatedAt.Seconds)},
	}
}

func (v *Values) PathString() string {
	v.init()
	if v.GetID() == "" {
		return v.GetType()
	}
	return fmt.Sprintf("%s/%s", v.GetType(), v.GetID())
}

// Set set an entry in the Node
func (v *Values) Set(k string, val *structpb.Value) {
	v.init()
	v.Fields[k] = val
}

// SetAll set all entries in the Node
func (v *Values) SetAll(data map[string]interface{}) {
	v.init()
	if data == nil {
		return
	}
	for k, val := range data {
		v.Set(k, ToValue(val))
	}
}

// Get gets an entry from the Node by key
func (v *Values) Get(key string) *structpb.Value {
	v.init()
	return v.Fields[key]
}

// Exists returns true if the key exists in the Node
func (v *Values) Exists(key string) bool {
	v.init()
	if val, ok := v.Fields[key]; ok && val != nil {
		return true
	}
	return false
}

// GetString gets an entry from the Node by key
func (v *Values) GetString(key string) string {
	v.init()
	if !v.Exists(key) {
		return ""
	}
	return v.Fields[key].GetStringValue()
}

func (v *Values) GetBool(key string) bool {
	if !v.Exists(key) {
		return false
	}
	return v.Fields[key].GetBoolValue()
}

func (v *Values) GetInt(key string) int {
	if !v.Exists(key) {
		return 0
	}
	return int(v.Fields[key].GetNumberValue())
}

func (v *Values) GetFloat(key string) float64 {
	if !v.Exists(key) {
		return 0
	}
	return v.Fields[key].GetNumberValue()
}

// Del deletes the entry from the Node by key
func (v *Values) Del(key string) {
	v.init()
	delete(v.Fields, key)
}

// Range iterates over the Node with the function. If the function returns false, the iteration exits.
func (v *Values) Range(iterator func(key string, val *structpb.Value) bool) {
	v.init()
	for k, val := range v.GetFields() {
		if !iterator(k, val) {
			break
		}
	}
}

// Filter returns a Node of the node that return true from the filter function
func (v *Values) Filter(filter func(key string, v *structpb.Value) bool) *Values {
	v.init()
	data := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	v.Range(func(key string, val *structpb.Value) bool {
		if filter(key, val) {
			v.Set(key, val)
		}
		return true
	})
	return &Values{data}
}

// Copy creates a replica of the Node
func (v *Values) Copy(values *structpb.Struct) *structpb.Struct {
	copied := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	if values == nil {
		return copied
	}
	v.Range(func(k string, val *structpb.Value) bool {
		v.Set(k, val)
		return true
	})
	return copied
}

func (v *Values) Equals(other *Values) bool {
	return reflect.DeepEqual(v, other)
}

func (v *Values) GetNested(key string) (*structpb.Struct, bool) {
	if val, ok := v.Fields[key]; ok && val != nil {
		if node := val.GetStructValue(); node != nil {
			return node, true
		}
	}
	return nil, false
}

func (v *Values) IsNested(key string) bool {
	_, ok := v.GetNested(key)
	return ok
}

func (v *Values) SetNested(key string, nested *structpb.Struct) {
	v.Set(key, ToValue(nested))
}

func ToMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case *structpb.Struct:
		vals := &Values{o}
		return vals.Map()
	case *Values:
		return o.Map()
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
	case map[string]interface{}:
		return o
	default:
		values := map[string]interface{}{}
		typeOf := reflect.TypeOf(o)
		valOf := reflect.ValueOf(o)
		for i := 0; i < typeOf.NumField(); i++ {
			field := typeOf.Field(i)
			values[field.Name] = valOf.Field(i).Interface()
		}
		return values
	}
}

func envFrom(obj interface{}) (*cel.Env, error) {
	var declarations []*exprpb.Decl
	data := ToMap(obj)
	for k, _ := range data {
		declarations = append(declarations, decls.NewVar(k, decls.Any))
	}
	return cel.NewEnv(cel.Declarations(declarations...))
}

func BooleanExpression(expressions []string, obj interface{}) (bool, error) {
	values := ToMap(obj)
	var programs []cel.Program
	env, err := envFrom(obj)
	if err != nil {
		return false, err
	}
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return false, iss.Err()
		}
		prgm, err := env.Program(ast)
		if err != nil {
			return false, err
		}
		programs = append(programs, prgm)
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(values)
		if err != nil {
			return false, err
		}
		if val, ok := out.Value().(bool); !ok || !val {
			passes = false
		}
	}
	return passes, nil
}

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
