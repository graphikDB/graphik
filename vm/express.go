package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

func init() {
	var err error
	anyMap := decls.NewMapType(decls.String, decls.Any)
	e, err = cel.NewEnv(cel.Declarations(
		decls.NewVar("path", anyMap),
		decls.NewVar("attributes", anyMap),
		decls.NewVar("from", anyMap),
		decls.NewVar("to", anyMap),
		decls.NewVar("cascade", decls.String),
		decls.NewVar("metadata", anyMap),
		decls.NewVar("channel", decls.String),
		decls.NewVar("method", decls.String),
		decls.NewVar("sender", anyMap),
		decls.NewVar("data", anyMap),
		decls.NewVar("timestamp", decls.Timestamp),
		decls.NewVar("paths", decls.NewListType(anyMap)),
		decls.NewVar("edges", decls.NewListType(anyMap)),
		decls.NewVar("nodes", decls.NewListType(anyMap)),
		decls.NewVar("edges_from", decls.NewListType(anyMap)),
		decls.NewVar("edges_to", decls.NewListType(anyMap)),
		decls.NewVar("request", anyMap),
		decls.NewVar("response", anyMap),
		decls.NewVar("limit", decls.Int),
		decls.NewVar("identity", anyMap),
		decls.NewVar("expressions", decls.NewListType(decls.String)),
		decls.NewVar("directed", decls.Bool),
		decls.NewVar("edge_changes", decls.NewListType(anyMap)),
		decls.NewVar("node_changes", decls.NewListType(anyMap)),
		/*
			"channel":   n.GetChannel(),
				"sender":    n.GetSender().AsMap(),
				"data":      n.Data.AsMap(),
				"timestamp": n.GetTimestamp(),
		*/
	))

	if err != nil {
		panic(err)
	}
}

var e *cel.Env

func Program(expression string) (cel.Program, error) {
	ast, iss := e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return e.Program(ast)
}

func Programs(expressions []string) ([]cel.Program, error) {
	var programs []cel.Program
	for _, exp := range expressions {
		prgm, err := Program(exp)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prgm)
	}
	return programs, nil
}

func Eval(programs []cel.Program, mapper apipb.Mapper) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(mapper.AsMap())
		if err != nil {
			return false, err
		}
		if val, ok := out.Value().(bool); !ok || !val {
			passes = false
		}
	}
	return passes, nil
}
