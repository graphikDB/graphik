package lang_test

import (
	"github.com/autom8ter/graphik/graph"
	"github.com/autom8ter/graphik/lang"
	"testing"
)

func Test(t *testing.T) {
	g := graph.New()
	vm := lang.NewVM(g)
	result, _, err := vm.Private().MapEval(`create_node(input)`, map[string]interface{}{
		"path": "user",
		"name": "tom",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
	result, _, err = vm.Private().MapEval(`create_node(input)`, map[string]interface{}{
		"path": "user",
		"name": "bob",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
	nodes := g.Nodes().All()
	for _, n := range nodes {
		t.Log(n)
	}
}
