package generic_test

import (
	"encoding/json"
	"github.com/autom8ter/graphik/generic"
	"github.com/autom8ter/graphik/graph/model"
	"testing"
)

var testNodes = []*model.Node{
	{
		Type: "user",
		Attributes: map[string]interface{}{
			"name": "bob",
		},
	},
	{
		ID:   "frd3",
		Type: "user",
		Attributes: map[string]interface{}{
			"name": "fred",
		},
	},
}

func TestNodes(t *testing.T) {
	nodes := generic.NewNodes()
	nodes.SetAll(testNodes...)
	results, err := nodes.Search("attributes.name", "user")
	if err != nil {
		t.Fatal(err)
	}
	bits, _ := json.Marshal(results)
	t.Log(string(bits))
}
