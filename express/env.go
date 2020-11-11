package express

import (
	"bytes"
	"encoding/json"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	var err error
	e, err = cel.NewEnv(cel.Declarations(
		decls.NewVar("gid", decls.String),
		decls.NewVar("gtype", decls.String),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("created_at", decls.Int),
		decls.NewVar("updated_at", decls.Int),
		decls.NewVar("from", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("to", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("cascade", decls.String),
	))
	if err != nil {
		panic(err)
	}
}

var e *cel.Env

func Eval(expressions []string, obj interface{}) (bool, error) {
	if len(expressions) == 0 || expressions[0] == "" {
		return true, nil
	}
	values := ToMap(obj)
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := e.Compile(exp)
		if iss.Err() != nil {
			return false, iss.Err()
		}
		prgm, err := e.Program(ast)
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

func ToMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case nil:
		return map[string]interface{}{}
	case map[string]interface{}:
		return o
	case *apipb.Node:
		return map[string]interface{}{
			"path": map[string]interface{}{
				"gid":   o.GetPath().GetGid(),
				"gtype": o.GetPath().GetGtype(),
			},
			"attributes": o.GetAttributes(),
			"created_at": o.GetCreatedAt(),
			"updated_at": o.GetUpdatedAt(),
		}
	case *apipb.Edge:
		return map[string]interface{}{
			"path": map[string]interface{}{
				"gid":   o.GetPath().GetGid(),
				"gtype": o.GetPath().GetGtype(),
			},
			"attributes": o.GetAttributes(),
			"cascade":    o.GetCascade().String(),
			"from": map[string]interface{}{
				"gid":  o.GetFrom().GetGid(),
				"type": o.GetFrom().GetGtype(),
			},
			"to": map[string]interface{}{
				"gid":   o.GetFrom().GetGid(),
				"gtype": o.GetFrom().GetGtype(),
			},
			"created_at": o.GetCreatedAt(),
			"updated_at": o.GetUpdatedAt(),
		}
	case *empty.Empty:
		return map[string]interface{}{}
	case *structpb.Struct:
		return o.AsMap()
	case string, int, int64, int32, bool:
		return map[string]interface{}{
			"value": o,
		}
	case proto.Message:
		buf := bytes.NewBuffer(nil)
		var data = map[string]interface{}{}
		jSONEncode(buf, o)
		json.Unmarshal(buf.Bytes(), &data)
		return data
	default:
		values := map[string]interface{}{}
		mapstructure.WeakDecode(o, &values)
		return values
	}
}
