package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type MessageVM struct {
	e *cel.Env
}

func NewMessageVM() (*MessageVM, error) {
	e, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("message", decls.NewMapType(decls.String, decls.Any)),
		),
	)
	if err != nil {
		return nil, err
	}
	return &MessageVM{e: e}, nil
}

func (n *MessageVM) Program(expression string) (cel.Program, error) {
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *MessageVM) Programs(expressions []string) ([]cel.Program, error) {
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

func (n *MessageVM) Eval(programs []cel.Program, message *apipb.Message) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"message": message.AsMap(),
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
