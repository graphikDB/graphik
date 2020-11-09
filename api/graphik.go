package apipb

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/hashicorp/raft"
	"github.com/mitchellh/mapstructure"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"reflect"
	"sort"
	"strings"
)

func toMap(obj interface{}) map[string]interface{} {
	switch o := obj.(type) {
	case *Node:
		return map[string]interface{}{
			"path":       toMap(o.Path),
			"attributes": o.Attributes,
			"updated_at": o.UpdatedAt,
			"created_at": o.CreatedAt,
		}
	case *Edge:
		return map[string]interface{}{
			"path":       toMap(o.Path),
			"attributes": o.Attributes,
			"updated_at": o.UpdatedAt,
			"created_at": o.CreatedAt,
			"from":       o.From,
			"to":         o.To,
			"cascade":    o.Cascade,
			"mutual":     o.Mutual,
		}
	case *Path:
		return map[string]interface{}{
			"id":   o.ID,
			"type": o.Type,
		}
	case *Edges:
		return map[string]interface{}{
			"edges": o.Edges,
		}
	case *Nodes:
		return map[string]interface{}{
			"nodes": o.Nodes,
		}
	case *structpb.Struct:
		return FromStruct(o)
	case proto.Message:
		values := map[string]interface{}{}
		buf := bytes.NewBuffer(nil)
		marshaller.Marshal(buf, o)
		json.Unmarshal(buf.Bytes(), &values)
		return values
	default:
		values := map[string]interface{}{}
		mapstructure.WeakDecode(o, &values)
		return values
	}
}

func envFrom(obj interface{}) (*cel.Env, error) {
	mapStrDyn := decls.NewMapType(decls.String, decls.Dyn)
	var declarations []*exprpb.Decl
	switch o := obj.(type) {
	case *Node, Node:
		declarations = append(declarations, decls.NewVar("path", mapStrDyn))
		declarations = append(declarations, decls.NewVar("attributes", mapStrDyn))
		declarations = append(declarations, decls.NewVar("created_at", decls.Timestamp))
		declarations = append(declarations, decls.NewVar("updated_at", decls.Timestamp))
	case *Edge, Edge:
		declarations = append(declarations, decls.NewVar("path", mapStrDyn))
		declarations = append(declarations, decls.NewVar("attributes", mapStrDyn))
		declarations = append(declarations, decls.NewVar("created_at", decls.Timestamp))
		declarations = append(declarations, decls.NewVar("updated_at", decls.Timestamp))
		declarations = append(declarations, decls.NewVar("from", mapStrDyn))
		declarations = append(declarations, decls.NewVar("to", mapStrDyn))
		declarations = append(declarations, decls.NewVar("mutual", decls.Bool))
		declarations = append(declarations, decls.NewVar("cascade", decls.String))
	case proto.Message:
		values := map[string]interface{}{}
		buf := bytes.NewBuffer(nil)
		marshaller.Marshal(buf, o)
		json.Unmarshal(buf.Bytes(), &values)
		for k, _ := range values {
			declarations = append(declarations, decls.NewVar(k, decls.Any))
		}
	default:
		val := reflect.ValueOf(o)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			declarations = append(declarations, decls.NewVar(field.String(), decls.Any))
		}
	}
	return cel.NewEnv(cel.Declarations(declarations...))
}

func EvaluateExpressions(expressions []string, obj interface{}) (bool, error) {
	values := toMap(obj)
	var programs []cel.Program
	env, err := envFrom(obj)
	if err != nil {
		return false, err
	}
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

func (e *Edges) Sort() {
	sorter := Sortable{
		LenFunc: func() int {
			if e == nil {
				return 0
			}
			return len(e.Edges)
		},
		LessFunc: func(i, j int) bool {
			if e == nil {
				return false
			}
			return e.Edges[i].UpdatedAt.Nanos > e.Edges[j].UpdatedAt.Nanos
		},
		SwapFunc: func(i, j int) {
			if e == nil {
				return
			}
			e.Edges[i], e.Edges[j] = e.Edges[j], e.Edges[i]
		},
	}
	sorter.Sort()
}

func (n *Nodes) Sort() {
	sorter := Sortable{
		LenFunc: func() int {
			if n == nil {
				return 0
			}
			return len(n.Nodes)
		},
		LessFunc: func(i, j int) bool {
			if n == nil {
				return false
			}
			return n.Nodes[i].UpdatedAt.Nanos > n.Nodes[j].UpdatedAt.Nanos
		},
		SwapFunc: func(i, j int) {
			if n == nil {
				return
			}
			n.Nodes[i], n.Nodes[j] = n.Nodes[j], n.Nodes[i]
		},
	}
	sorter.Sort()
}

func (c *Config) SetDefaults() {
	if c.Grpc == nil {
		c.Grpc = &GRPCConfig{}
	}
	if c.Grpc.Bind == "" {
		c.Grpc.Bind = ":7820"
	}
	if c.Http == nil {
		c.Http = &HTTPConfig{}
	}
	if c.Http.Bind == "" {
		c.Http.Bind = ":7830"
	}
	if c.Raft == nil {
		c.Raft = &RaftConfig{}
	}
	if c.Raft.Bind == "" {
		c.Raft.Bind = "localhost:7840"
	}
	if c.Raft.StoragePath == "" {
		c.Raft.StoragePath = "/tmp/graphik"
	}
	if c.Raft.NodeId == "" {
		c.Raft.NodeId = "default"
	}
}

type Sortable struct {
	LenFunc  func() int
	LessFunc func(i, j int) bool
	SwapFunc func(i, j int)
}

func (s Sortable) Len() int {
	if s.LenFunc == nil {
		return 0
	}
	return s.LenFunc()
}

func (s Sortable) Less(i, j int) bool {
	if s.LessFunc == nil {
		return false
	}
	return s.LessFunc(i, j)
}

func (s Sortable) Swap(i, j int) {
	if s.SwapFunc == nil {
		return
	}
	s.SwapFunc(i, j)
}

func (s Sortable) Sort() {
	sort.Sort(s)
}
