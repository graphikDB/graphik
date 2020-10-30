package generic_test

import (
	"github.com/autom8ter/graphik/graph/model"
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
