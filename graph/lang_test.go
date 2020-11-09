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
			"_id": "tom2",
			"_type": "user",
			"name":  "tom",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	result, _, err = g.Expression(`create_node(node)`, graph.NewValues(map[string]interface{}{
		"node": map[string]interface{}{
			"_id": "bob3",
			"_type": "user",
			"name":  "bob",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	result, _, err = g.Expression(`get_nodes(_type)`, map[string]interface{}{
		"_type": "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprint(result.Value()))
}
