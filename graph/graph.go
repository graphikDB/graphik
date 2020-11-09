package graph

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/pkg/errors"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type Graph struct {
	nodes *NodeStore
	edges *EdgeStore
}

func New() *Graph {
	edges := newEdgeStore()
	return &Graph{
		nodes: newNodeStore(edges),
		edges: edges,
	}
}

func (g *Graph) Edges() *EdgeStore {
	return g.edges
}

func (g *Graph) Nodes() *NodeStore {
	return g.nodes
}

func (g *Graph) Close() {
	g.edges.Close()
	g.nodes.Close()
}

func (g *Graph) Do(fn func(g *Graph)) {
	fn(g)
}

func (g *Graph) Functions() []*functions.Overload {
	return []*functions.Overload{
		{
			Operator: "create_node_map_map",
			Unary: func(lhs ref.Val) ref.Val {
				vals := ToMap(lhs.Value())
				return types.NewDynamicMap(types.NewRegistry(), g.nodes.Set(vals))
			},
		},
		{
			Operator: "create_edge_map_map",
			Unary: func(lhs ref.Val) ref.Val {
				vals := ToMap(lhs.Value())
				return types.NewDynamicMap(types.NewRegistry(), g.edges.Set(vals))
			},
		},
		{
			Operator: "get_node_string_map",
			Unary: func(lhs ref.Val) ref.Val {
				val, ok := g.nodes.Get(lhs.Value().(string))
				if !ok {
					return types.NullType
				}
				return types.NewDynamicMap(types.NewRegistry(), val)
			},
		},
		{
			Operator: "get_nodes_map_map",
			Unary: func(lhs ref.Val) ref.Val {
				m := ToMap(lhs.Value())
				values, err := g.nodes.FilterSearch(FilterFrom(m))
				if err != nil {
					panic(err)
				}
				return types.NewDynamicMap(types.NewRegistry(), map[string]interface{}{
					"nodes": values,
				})
			},
		},
		{
			Operator: "get_edges_map_map",
			Unary: func(lhs ref.Val) ref.Val {
				m := ToMap(lhs.Value())
				values, err := g.edges.FilterSearch(FilterFrom(m))
				if err != nil {
					panic(err)
				}
				return types.NewDynamicMap(types.NewRegistry(), map[string]interface{}{
					"edges": values,
				})
			},
		},
		{
			Operator: "edges_from_string_string_map",
			Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
				m := ToMap(rhs.Value())
				values := g.edges.RangeFilterFrom(lhs.Value().(string), &Filter{
					Type:        m["type"].(string),
					Expressions: m["expressions"].([]string),
					Limit:       m["limit"].(int),
				})
				return types.NewDynamicList(types.NewRegistry(), values)
			},
		},
		{
			Operator: "edges_to_string_string_map",
			Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
				m := ToMap(rhs.Value())
				values := g.edges.RangeFilterTo(lhs.Value().(string), &Filter{
					Type:        m["type"].(string),
					Expressions: m["expressions"].([]string),
					Limit:       m["limit"].(int),
				})
				return types.NewDynamicList(types.NewRegistry(), values)
			},
		},
	}
}

func (g *Graph) Env(obj interface{}) (*cel.Env, error) {
	var declarations = []*exprpb.Decl{
		decls.NewFunction("create_node",
			decls.NewOverload("create_node_map_map",
				[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
				decls.NewMapType(decls.String, decls.Any))),
		decls.NewFunction("create_edge",
			decls.NewOverload("create_edge_map_map",
				[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
				decls.NewMapType(decls.String, decls.Any))),
		decls.NewFunction("get_node",
			decls.NewOverload("get_node_string_map",
				[]*exprpb.Type{decls.String},
				decls.NewMapType(decls.String, decls.Any))),
		decls.NewFunction("get_nodes",
			decls.NewOverload("get_nodes_map_map",
				[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
				decls.NewListType(decls.NewMapType(decls.String, decls.Any)))),
		decls.NewFunction("get_edges",
			decls.NewOverload("get_edges_map_map",
				[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
				decls.NewListType(decls.NewMapType(decls.String, decls.Any)))),
		decls.NewFunction("edges_from",
			decls.NewOverload("edges_from_string_string_map",
				[]*exprpb.Type{decls.String, decls.NewMapType(decls.String, decls.Any)},
				decls.NewListType(decls.NewMapType(decls.String, decls.Any)))),
		decls.NewFunction("edges_to",
			decls.NewOverload("edges_to_string_string_map",
				[]*exprpb.Type{decls.String, decls.NewMapType(decls.String, decls.Any)},
				decls.NewListType(decls.NewMapType(decls.String, decls.Any)))),
	}
	data := ToMap(obj)
	for k, _ := range data {
		declarations = append(declarations, decls.NewVar(k, decls.Any))
	}
	return cel.NewEnv(cel.Declarations(declarations...))
}

func (g *Graph) Expression(expression string, obj interface{}) (ref.Val, *cel.EvalDetails, error) {
	values := ToMap(obj)
	env, err := g.Env(obj)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating env")
	}
	ast, iss := env.Compile(expression)
	if iss.Err() != nil {
		return nil, nil, errors.Wrap(iss.Err(), "compiling env")
	}

	program, err := env.Program(ast, cel.Functions(g.Functions()...))
	if err != nil {
		return nil, nil, errors.Wrap(err, "programming env")
	}
	x, y, err := program.Eval(values)
	if err != nil {
		return nil, nil, errors.Wrap(err, "evaluating program")
	}
	return x, y, nil
}

func (g *Graph) BooleanExpression(expressions []string, obj interface{}) (bool, error) {
	if len(expressions) == 0 {
		return true, nil
	}
	values := ToMap(obj)
	env, err := g.Env(obj)
	if err != nil {
		return false, err
	}
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return false, iss.Err()
		}
		prgm, err := env.Program(ast, cel.Functions(g.Functions()...))
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
