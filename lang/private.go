package lang

import (
	"github.com/autom8ter/graphik/graph"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func privateFuncMap(g *graph.Graph) FuncMap {
	return map[string]*Function{
		"create_node": {
			Overload: &functions.Overload{
				Operator: "create_node_map_map",
				Unary: func(lhs ref.Val) ref.Val {
					vals := graph.ToMap(lhs.Value())
					return types.NewDynamicMap(types.NewRegistry(), g.Nodes().Set(vals))
				},
			},
			Declaration: &Declaration{
				Name: "create_node",
				Overloads: []*exprpb.Decl_FunctionDecl_Overload{
					decls.NewOverload("create_node_map_map",
						[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
						decls.NewMapType(decls.String, decls.Any)),
				},
			},
		},
	}
}
