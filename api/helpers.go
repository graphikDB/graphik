package apipb

import (
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/struct"
	"io"
	"reflect"
)

var (
	marshaller = &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "",
		OrigName:     false,
		AnyResolver:  nil,
	}
)

func JSONEncode(w io.Writer, msg proto.Message) error {
	return marshaller.Marshal(w, msg)
}

func JSONDecode(r io.Reader, msg proto.Message) error {
	return jsonpb.Unmarshal(r, msg)
}

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
