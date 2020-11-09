package graph_test

import (
	"github.com/autom8ter/graphik/graph"
	"testing"
)

func Test(t *testing.T) {
	val, err := graph.BooleanExpression([]string{`_type.startsWith("user")`}, graph.NewValues(map[string]interface{}{
		"_type": "user",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !val {
		t.Fatal("failure")
	}
}
