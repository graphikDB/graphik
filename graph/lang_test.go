package graph_test

import (
	"fmt"
	"github.com/autom8ter/graphik/graph"
	"testing"
)

func Test(t *testing.T) {
	g := graph.New()
	result, _, err := g.Expression(`create_node(node)`, graph.NewValues(map[string]interface{}{
		"node": map[string]interface{}{
			"path": "user",
			"name":  "tom",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	result, _, err = g.Expression(`create_node(node)`, graph.NewValues(map[string]interface{}{
		"node": map[string]interface{}{
			"path": "user",
			"name":  "bob",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	nodes := g.Nodes().All()
	for _, n := range nodes {
		t.Log(n)
	}
	result, _, err = g.Expression(`get_nodes(filter)`, map[string]interface{}{
		"filter": map[string]interface{}{
			"type":       "*",
			"expressions": []string{},
			"limit":       5,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprint(result.Value()))
}
