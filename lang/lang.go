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
	idKey        = "_id"
	typeKey      = "_type"
	createdAtKey = "_createdAt"
	updatedAtKey = "_updatedAt"
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

// ToStruct converts a map[string]interface{} to a ptypes.Struct
func ToStruct(v map[string]interface{}) *structpb.Struct {
	size := len(v)
	if size == 0 {
		return nil
	}
	fields := make(map[string]*structpb.Value, size)
	for k, v := range v {
		fields[k] = ToValue(v)
	}
	return &structpb.Struct{
		Fields: fields,
	}
}

func FromStruct(s *structpb.Struct) map[string]interface{} {
	values := map[string]interface{}{}
	for k, field := range s.Fields {
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
		return FromStruct(field.GetStructValue())
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

func initStruct(values *structpb.Struct) {
	if values == nil {
		values = &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		}
	}
}

func GetType(values *structpb.Struct) string {
	initStruct(values)
	return values.GetFields()[typeKey].GetStringValue()
}

func GetID(values *structpb.Struct) string {
	initStruct(values)
	return values.GetFields()[idKey].GetStringValue()
}

func GetCreatedAt(values *structpb.Struct) int64 {
	initStruct(values)
	return int64(values.GetFields()[createdAtKey].GetNumberValue())
}

func GetUpdatedAt(values *structpb.Struct) int64 {
	initStruct(values)
	return int64(values.GetFields()[updatedAtKey].GetNumberValue())
}

func SetID(values *structpb.Struct, _id string) {
	initStruct(values)
	values.GetFields()[idKey] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: _id},
	}
}

func SetType(values *structpb.Struct, _type string) {
	initStruct(values)
	values.GetFields()[typeKey] = &structpb.Value{
		Kind: &structpb.Value_StringValue{StringValue: _type},
	}
}

func SetCreatedAt(values *structpb.Struct, _createdAt *timestamp.Timestamp) {
	initStruct(values)
	values.GetFields()[createdAtKey] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(_createdAt.Seconds)},
	}
}

func SetUpdatedAt(values *structpb.Struct, _updatedAt *timestamp.Timestamp) {
	initStruct(values)
	values.GetFields()[updatedAtKey] = &structpb.Value{
		Kind: &structpb.Value_NumberValue{NumberValue: float64(_updatedAt.Seconds)},
	}
}

func PathString(values *structpb.Struct) string {
	initStruct(values)
	if GetID(values) == "" {
		return GetType(values)
	}
	return fmt.Sprintf("%s/%s", GetType(values), GetID(values))
}

// Set set an entry in the Node
func Set(values *structpb.Struct, k string, val *structpb.Value) {
	initStruct(values)
	values.Fields[k] = val
}

// SetAll set all entries in the Node
func SetAll(values *structpb.Struct, data map[string]interface{}) {
	initStruct(values)
	if data == nil {
		return
	}
	for k, val := range data {
		Set(values, k, ToValue(val))
	}
}

// Get gets an entry from the Node by key
func Get(values *structpb.Struct, key string) *structpb.Value {
	initStruct(values)
	return values.Fields[key]
}

// Exists returns true if the key exists in the Node
func Exists(values *structpb.Struct, key string) bool {
	initStruct(values)
	if val, ok := values.Fields[key]; ok && val != nil {
		return true
	}
	return false
}

// GetString gets an entry from the Node by key
func GetString(values *structpb.Struct, key string) string {
	initStruct(values)
	if !Exists(values, key) {
		return ""
	}
	return values.Fields[key].GetStringValue()
}

func GetBool(values *structpb.Struct, key string) bool {
	if !Exists(values, key) {
		return false
	}
	return values.Fields[key].GetBoolValue()
}

func GetInt(values *structpb.Struct, key string) int {
	if !Exists(values, key) {
		return 0
	}
	return int(values.Fields[key].GetNumberValue())
}

func GetFloat(values *structpb.Struct, key string) float64 {
	if !Exists(values, key) {
		return 0
	}
	return values.Fields[key].GetNumberValue()
}

// Del deletes the entry from the Node by key
func Del(values *structpb.Struct, key string) {
	initStruct(values)
	delete(values.Fields, key)
}

// Range iterates over the Node with the function. If the function returns false, the iteration exits.
func Range(values *structpb.Struct, iterator func(key string, val *structpb.Value) bool) {
	initStruct(values)
	for k, val := range values.GetFields() {
		if !iterator(k, val) {
			break
		}
	}
}

// Filter returns a Node of the node that return true from the filter function
func Filter(values *structpb.Struct, filter func(key string, v *structpb.Value) bool) *structpb.Struct {
	data := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	if values == nil {
		return data
	}
	Range(values, func(key string, val *structpb.Value) bool {
		if filter(key, val) {
			Set(values, key, val)
		}
		return true
	})
	return data
}

// Copy creates a replica of the Node
func Copy(values *structpb.Struct) *structpb.Struct {
	copied := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	if values == nil {
		return copied
	}
	Range(values, func(k string, val *structpb.Value) bool {
		Set(values, k, val)
		return true
	})
	return copied
}

func Equals(values *structpb.Struct, other *structpb.Struct) bool {
	return reflect.DeepEqual(values, other)
}

func GetNested(values *structpb.Struct, key string) (*structpb.Struct, bool) {
	if val, ok := values.Fields[key]; ok && val != nil {
		if node := val.GetStructValue(); node != nil {
			return node, true
		}
	}
	return nil, false
}

func IsNested(values *structpb.Struct, key string) bool {
	_, ok := GetNested(values, key)
	return ok
}

func SetNested(values *structpb.Struct, key string, nested *structpb.Struct) {
	Set(values, key, ToValue(nested))
}

func ToMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case *structpb.Struct:
		return FromStruct(o)
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
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

