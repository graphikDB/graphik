package model

import (
	"encoding/json"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/mitchellh/mapstructure"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"io"
	"strconv"
	"strings"
	"time"
)

func Decode(input, output interface{}) error {
	return mapstructure.Decode(input, output)
}

type Path struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *Path) String() string {
	return fmt.Sprintf("%s/%s", p.Type, p.ID)
}

func PathFromString(path string) Path {
	parts := strings.Split(path, "/")
	if len(parts) == 2 {
		return Path{
			ID:   parts[1],
			Type: parts[0],
		}
	}
	return Path{
		Type: parts[0],
	}
}

func (p *Path) UnmarshalGQL(v interface{}) error {
	pointStr, ok := v.(string)
	if !ok {
		return fmt.Errorf("path must be string ({type}/{id})")
	}
	parts := strings.Split(pointStr, "/")
	if len(parts) == 0 {
		return fmt.Errorf("empty path ({type}/{id})")
	}
	if len(parts) > 2 {
		return fmt.Errorf("path contains multiple separators(/) %s ({type}/{id})", pointStr)
	}
	p.Type = parts[0]
	if p.Type == "" {
		return fmt.Errorf("path does not contain type ({type}/{id})")
	}
	if len(parts) == 2 {
		p.ID = parts[1]
	}
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (p Path) MarshalGQL(w io.Writer) {
	json.NewEncoder(w).Encode(p.String())
}

type Node struct {
	Path       Path                   `json:"path"`
	Attributes map[string]interface{} `json:"attributes"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

func (n *Node) Map() map[string]interface{} {
	return map[string]interface{}{
		"path":       n.Path.String(),
		"attributes": n.Attributes,
		"createdAt":  n.CreatedAt,
		"updatedAt":  n.UpdatedAt,
	}
}

type Edge struct {
	Path       Path                   `json:"path"`
	Mutual     bool                   `json:"mutual"`
	Attributes map[string]interface{} `json:"attributes"`
	From       Path                   `json:"from"`
	To         Path                   `json:"to"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

func (e *Edge) Map() map[string]interface{} {
	return map[string]interface{}{
		"path":       e.Path.String(),
		"attributes": e.Attributes,
		"mutual":     e.Mutual,
		"from": map[string]interface{}{
			"id":   e.From.ID,
			"type": e.From.Type,
		},
		"to": map[string]interface{}{
			"id":   e.To.ID,
			"type": e.To.Type,
		},
		"createdAt": e.CreatedAt,
		"updatedAt": e.UpdatedAt,
	}
}

type Mapper interface {
	Map() map[string]interface{}
}

func nodeDeclarations() []*exprpb.Decl {
	return []*exprpb.Decl{
		decls.NewVar("path", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("createdAt", decls.Timestamp),
		decls.NewVar("updatedAt", decls.Timestamp),
	}
}

func edgeDeclarations() []*exprpb.Decl {
	n := nodeDeclarations()
	n = append(n,
		decls.NewVar("mutual", decls.Bool),
		decls.NewVar("from", decls.NewMapType(decls.String, decls.Any)),
		decls.NewVar("to", decls.NewMapType(decls.String, decls.Any)),
	)
	return n
}

func Evaluate(expressions []string, mapper Mapper) (bool, error) {
	var declarations []*exprpb.Decl
	if _, ok := mapper.(*Node); ok {
		declarations = nodeDeclarations()
	}
	if _, ok := mapper.(*Edge); ok {
		declarations = edgeDeclarations()
	}
	env, err := cel.NewEnv(cel.Declarations(declarations...))
	if err != nil {
		return false, err
	}
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return false, err
		}
		prgm, err := env.Program(ast)
		if err != nil {
			return false, err
		}
		programs = append(programs, prgm)
	}
	var passes = true
	for _, program := range programs {
		out, _, err := program.Eval(mapper.Map())
		if err != nil {
			return false, err
		}
		if val, ok := out.Value().(bool); !ok || !val {
			passes = false
		}
	}
	return passes, nil
}

type Message struct {
	Channel string                 `json:"channel"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data"`
}

func (m *Message) UnmarshalGQL(v interface{}) error {
	return Decode(v, m)
}

// MarshalGQL implements the graphql.Marshaler interface
func (m Message) MarshalGQL(w io.Writer) {
	json.NewEncoder(w).Encode(&m)
}

type ChangeFilter struct {
	Op          Op       `json:"op"`
	Type        string   `json:"type"`
	Expressions []string `json:"expressions"`
}

type Command struct {
	Op        Op          `json:"op"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type Counter struct {
	Count int `json:"count"`
}

type DepthFilter struct {
	Depth       int      `json:"depth"`
	Path        Path     `json:"path"`
	Expressions []string `json:"expressions"`
	Limit       int      `json:"limit"`
	Reverse     *bool    `json:"reverse"`
}

type EdgeConstructor struct {
	Path       Path                   `json:"path"`
	Mutual     bool                   `json:"mutual"`
	Attributes map[string]interface{} `json:"attributes"`
	From       Path                   `json:"from"`
	To         Path                   `json:"to"`
}

type Filter struct {
	Type        string   `json:"type"`
	Expressions []string `json:"expressions"`
	Limit       int      `json:"limit"`
}

type NodeConstructor struct {
	Path       Path                   `json:"path"`
	Attributes map[string]interface{} `json:"attributes"`
}

type Export struct {
	Nodes []*Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

type Patch struct {
	Path  Path                   `json:"path"`
	Patch map[string]interface{} `json:"patch"`
}

type Op string

const (
	OpCreateNode Op = "CREATE_NODE"
	OpPatchNode  Op = "PATCH_NODE"
	OpDeleteNode Op = "DELETE_NODE"
	OpCreateEdge Op = "CREATE_EDGE"
	OpPatchEdge  Op = "PATCH_EDGE"
	OpDeleteEdge Op = "DELETE_EDGE"
)

var AllOp = []Op{
	OpCreateNode,
	OpPatchNode,
	OpDeleteNode,
	OpCreateEdge,
	OpPatchEdge,
	OpDeleteEdge,
}

func (e Op) IsValid() bool {
	switch e {
	case OpCreateNode, OpPatchNode, OpDeleteNode, OpCreateEdge, OpPatchEdge, OpDeleteEdge:
		return true
	}
	return false
}

func (e Op) String() string {
	return string(e)
}

func (e *Op) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Op(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Op", str)
	}
	return nil
}

func (e Op) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Operator string

const (
	OperatorNeq Operator = "NEQ"
	OperatorEq  Operator = "EQ"
)

var AllOperator = []Operator{
	OperatorNeq,
	OperatorEq,
}

func (e Operator) IsValid() bool {
	switch e {
	case OperatorNeq, OperatorEq:
		return true
	}
	return false
}

func (e Operator) String() string {
	return string(e)
}

func (e *Operator) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Operator(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Operator", str)
	}
	return nil
}

func (e Operator) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
