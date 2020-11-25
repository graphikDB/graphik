package vm

import (
	"fmt"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type EdgeVM struct {
	e *cel.Env
}

func NewEdgeVM() (*EdgeVM, error) {
	fmt.Println(_edge.ProtoReflect().Descriptor().FullName())
	e, err := cel.NewEnv(
		cel.Types(
			_structp,
			_node,
			_path,
			_meta,
			_edge,
		),
		cel.Declarations(
			decls.NewVar("edge", decls.NewObjectType(string(_edge.ProtoReflect().Descriptor().FullName()))),
		),
	)
	if err != nil {
		return nil, err
	}
	return &EdgeVM{e: e}, nil
}

func (n *EdgeVM) Program(expression string) (cel.Program, error) {
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *EdgeVM) Programs(expressions []string) ([]cel.Program, error) {
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

func (n *EdgeVM) Eval(programs []cel.Program, edge *apipb.Edge) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"edge": edge,
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
