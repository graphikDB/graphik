package vm

import (
	"errors"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/graphikDB/graphik/gen/grpc/go"
	"strings"
)

type AuthVM struct {
	e *cel.Env
}

func NewAuthVM() (*AuthVM, error) {
	e, err := cel.NewEnv(
		cel.Types(&apipb.Doc{}, &apipb.Ref{}, &apipb.Request{}),
		cel.Declarations(
			decls.NewVar("this", decls.NewMapType(decls.String, decls.Any)),
		),
	)
	if err != nil {
		return nil, err
	}
	return &AuthVM{e: e}, nil
}

func (n *AuthVM) Program(expression string) (cel.Program, error) {
	if expression == "" {
		return nil, errors.New("empty expression")
	}
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *AuthVM) Programs(expressions []string) ([]cel.Program, error) {
	var programs []cel.Program
	for _, exp := range expressions {
		if exp == "" {
			continue
		}
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
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"this": req,
		})
		if err != nil {
			if strings.Contains(err.Error(), "no such key") {
				return false, nil
			}
			return false, err
		}
		if val, ok := out.Value().(bool); ok {
			if val {
				return true, nil
			}
		}
	}
	return false, nil
}
