package vm

import (
	"github.com/autom8ter/graphik/helpers"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

func init() {
	var err error
	e, err = cel.NewEnv(cel.Declarations(
		decls.NewVar("path", decls.NewMapType(decls.String, decls.Any)),
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

func Eval(programs []cel.Program, obj interface{}) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	values := helpers.ToMap(obj)
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
