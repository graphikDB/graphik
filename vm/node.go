package vm

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
)

type NodeVM struct {
	e *cel.Env
}

func NewNodeVM() (*NodeVM, error) {
	e, err := cel.NewEnv(
		cel.Types(
			_node,
			_path,
			_structp,
			_meta,
		),
		cel.Declarations(
			decls.NewVar("node", decls.NewObjectType(string(_node.ProtoReflect().Descriptor().FullName()))),
		),
	)
	if err != nil {
		return nil, err
	}
	return &NodeVM{e: e}, nil
}

func (n *NodeVM) Program(expression string) (cel.Program, error) {
	ast, iss := n.e.Compile(expression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	return n.e.Program(ast)
}

func (n *NodeVM) Programs(expressions []string) ([]cel.Program, error) {
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

func (n *NodeVM) Eval(programs []cel.Program, node *apipb.Node) (bool, error) {
	if len(programs) == 0 || programs[0] == nil {
		return true, nil
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(map[string]interface{}{
			"node": node,
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
