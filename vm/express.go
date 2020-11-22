package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

func init() {
	var err error
	e, err = cel.NewEnv(cel.Declarations(
		decls.NewVar("path", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("from", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("to", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("cascade", decls.String),
		decls.NewVar("metadata", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("channel", decls.String),
		decls.NewVar("sender", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("data", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("timestamp", decls.Timestamp),
		decls.NewVar("paths", decls.NewListType(decls.NewMapType(decls.String, decls.Any))),
		decls.NewVar("edges", decls.NewListType(decls.NewMapType(decls.String, decls.Any))),
		decls.NewVar("nodes", decls.NewListType(decls.NewMapType(decls.String, decls.Any))),
		decls.NewVar("edges_from", decls.NewListType(decls.NewMapType(decls.String, decls.Any))),
		decls.NewVar("edges_to", decls.NewListType(decls.NewMapType(decls.String, decls.Any))),
		decls.NewVar("request", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("limit", decls.Int),
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
