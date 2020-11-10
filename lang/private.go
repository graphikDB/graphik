package lang

import (
	"github.com/autom8ter/graphik/graph"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/pkg/errors"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func privateFuncMap(g *graph.Graph) FuncMap {
	return map[string]*Function{
		"createNode": {
			Overload: &functions.Overload{
				Operator: "createNode",
				Unary: func(lhs ref.Val) ref.Val {
					vals := graph.ToMap(lhs.Value())
					return types.NewDynamicMap(types.NewRegistry(), g.Nodes().Set(vals))
				},
			},
			Declarations: []*exprpb.Decl_FunctionDecl_Overload{
				decls.NewOverload("createNode",
					[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
					decls.NewMapType(decls.String, decls.Any)),
			},
		},
		"getNode": {
			Overload: &functions.Overload{
				Operator: "getNode",
				Unary: func(lhs ref.Val) ref.Val {
					vals, ok := lhs.Value().(map[string]interface{})
					if !ok {
						return errVal(errors.New("input must map containing path field"))
					}
					if vals["path"] == nil {
						return errVal(errors.New("input must map containing path field"))
					}
					n, ok := g.Nodes().Get(vals["path"].(string))
					if !ok {
						return types.NewDynamicMap(types.NewRegistry(), errVal(errors.Errorf("%s not found", vals["path"].(string))))
					}
					return types.NewDynamicMap(types.NewRegistry(), n)
				},
			},
			Declarations: []*exprpb.Decl_FunctionDecl_Overload{
				decls.NewOverload("getNode",
					[]*exprpb.Type{decls.NewMapType(decls.String, decls.Any)},
					dynamic),
			},
		},
	}
}
