package generic

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

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

func filterNodePrograms(expressions []string) ([]cel.Program, error) {
	env, err := cel.NewEnv(cel.Declarations(nodeDeclarations()...))
	if err != nil {
		return nil, err
	}
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return nil, err
		}
		prgm, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prgm)
	}
	return programs, nil
}

func filterEdgePrograms(expressions []string) ([]cel.Program, error) {
	env, err := cel.NewEnv(cel.Declarations(edgeDeclarations()...))
	if err != nil {
		return nil, err
	}
	var programs []cel.Program
	for _, exp := range expressions {
		ast, iss := env.Compile(exp)
		if iss.Err() != nil {
			return nil, err
		}
		prgm, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		programs = append(programs, prgm)
	}
	return programs, nil
}
