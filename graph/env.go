package graph

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

func GetEnv() (*cel.Env, error) {
	return cel.NewEnv(cel.Declarations(
		decls.NewVar("gid", decls.String),
		decls.NewVar("gtype", decls.String),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("created_at", decls.Int),
		decls.NewVar("updated_at", decls.Int),
		decls.NewVar("from", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("to", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("cascade", decls.String),
	))
}

func EvalExpression(env *cel.Env, expressions []string, obj interface{}) (bool, error) {
	if len(expressions) == 0 {
		return true, nil
	}
	values := ToMap(obj)
	var programs []cel.Program
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
