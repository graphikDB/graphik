package apipb

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"strings"
)

func init() {
	var err error
	env, err = cel.NewEnv(cel.Types(&Node{}, &Edge{}, &Path{}, &Message{}))
	if err != nil {
		panic(err)
	}
}

var env *cel.Env

func EvaluateExpressions(expressions []string, obj interface{}) (bool, error) {
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
		out, _, err := program.Eval(obj)
		if err != nil {
			return false, err
		}
		if val, ok := out.Value().(bool); !ok || !val {
			passes = false
		}
	}
	return passes, nil
}

func (p *Path) PathString() string {
	if p.ID == "" {
		return p.Type
	}
	return fmt.Sprintf("%s/%s", p.Type, p.ID)
}

func PathFromString(path string) *Path {
	parts := strings.Split(path, "/")
	if len(parts) == 2 {
		return &Path{
			ID:   parts[1],
			Type: parts[0],
		}
	}
	return &Path{
		Type: parts[0],
	}
}
