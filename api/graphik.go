package apipb

import (
	"crypto/rand"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/cel-go/cel"
	"github.com/hashicorp/raft"
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

func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (c *Command) Log() *raft.Log {
	bits, _ := proto.Marshal(c)
	return &raft.Log{
		Data: bits,
	}
}
