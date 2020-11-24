package vm

import (
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type AuthVM struct {
	e *cel.Env
}

func NewAuthVM() (*AuthVM, error) {
	e, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("request", decls.NewMapType(decls.String, decls.Any)),
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

func (n *AuthVM) Eval(req *apipb.Request, programs ...cel.Program) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"request": map[string]interface{}{
				"method":    req.GetMethod(),
				"request":   req.GetRequest().AsMap(),
				"identity":  req.GetIdentity().AsMap(),
				"timestamp": req.GetTimestamp().Seconds,
			},
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
