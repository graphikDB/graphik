package vm

import (
	"errors"
	"github.com/autom8ter/graphik/gen/go/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type ChangeVM struct {
	e *cel.Env
}

func NewChangeVM() (*ChangeVM, error) {
	e, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("change", decls.NewMapType(decls.String, decls.Any)),
		),
	)
	if err != nil {
		return nil, err
	}
	return &ChangeVM{e: e}, nil
}

func (n *ChangeVM) Program(expression string) (cel.Program, error) {
	if expression == "" {
		return nil, errors.New("empty expression")
	}
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *ChangeVM) Programs(expressions []string) ([]cel.Program, error) {
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

func (n *ChangeVM) Eval(change *apipb.Change, programs ...cel.Program) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"change": change.AsMap(),
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
