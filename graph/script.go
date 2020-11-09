package graph

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types/ref"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func envFrom(obj interface{}) (*cel.Env, error) {
	var declarations []*exprpb.Decl
	data := ToMap(obj)
	for k, _ := range data {
		declarations = append(declarations, decls.NewVar(k, decls.Any))
	}
	return cel.NewEnv(cel.Declarations(declarations...))
}

func Expression(expression string, obj interface{}) (ref.Val, *cel.EvalDetails, error) {
	values := ToMap(obj)
	env, err := envFrom(obj)
	if err != nil {
		return nil, nil, err
	}
	ast, iss := env.Compile(expression)
	if iss.Err() != nil {
		return nil, nil, iss.Err()
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, nil, err
	}
	return program.Eval(values)
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
