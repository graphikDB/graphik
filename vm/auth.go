package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type AuthVM struct {
	e *cel.Env
}

func NewAuthVM() (*AuthVM, error) {
	e, err := cel.NewEnv(
		cel.Types(
			_node,
			_path,
			_meta,
		),
		cel.Declarations(
			decls.NewVar("request", decls.NewObjectType(string(_request.ProtoReflect().Descriptor().FullName()))),
		),
	)
	if err != nil {
		return nil, err
	}
	return &AuthVM{e: e}, nil
}

func (n *AuthVM) Program(expression string) (cel.Program, error) {
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *AuthVM) Programs(expressions []string) ([]cel.Program, error) {
	var programs []cel.Program
	for _, exp := range expressions {
		prgm, err := n.Program(exp)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prgm)
	}
	return programs, nil
}

func (n *AuthVM) Eval(programs []cel.Program, req *apipb.Request) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"request": req,
		})
		if err != nil {
			return false, err
		}
		if val, ok := out.Value().(bool); !ok || !val {
			passes = false
		}
	}
	return passes, nil
}
