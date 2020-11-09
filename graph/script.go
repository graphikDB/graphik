package graph

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func (g *Graph) Functions() []*functions.Overload {
	return []*functions.Overload{
		{
			Operator: "create_node_map_map",
			Unary: func(lhs ref.Val) ref.Val {
				m := ToMap(lhs.Value())
				return types.NewDynamicMap(types.NewRegistry(), g.nodes.Set(m))
			},
		},
	}
}

func (g *Graph) Env(obj interface{}) (*cel.Env, error) {
	var declarations = []*exprpb.Decl{
		decls.NewFunction("create_node",
			decls.NewOverload("create_node_map_map",
				[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
				decls.NewMapType(decls.String, decls.Any))),
	}
	data := ToMap(obj)
	for k, _ := range data {
		declarations = append(declarations, decls.NewVar(k, decls.Any))
	}
	return cel.NewEnv(cel.Declarations(declarations...))
}

func (g *Graph) Expression(expression string, obj interface{}) (ref.Val, *cel.EvalDetails, error) {
	values := ToMap(obj)
	env, err := g.Env(obj)
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

func (g *Graph) BooleanExpression(expressions []string, obj interface{}) (bool, error) {
	values := ToMap(obj)
	env, err := g.Env(obj)
	if err != nil {
		return false, err
	}
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return false, iss.Err()
		}
		prgm, err := env.Program(ast, cel.Functions(g.Functions()...))
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
	env, err := envFrom(obj)
	if err != nil {
		return false, err
	}
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
