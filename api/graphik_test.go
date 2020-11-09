package apipb_test

import (
	apipb "github.com/autom8ter/graphik/api"
	"testing"
)

func Test(t *testing.T) {
	val, err := apipb.EvaluateExpressions([]string{`path.type.startsWith("user")`}, &apipb.Node{
		Path: &apipb.Path{
			Type: "user",
		},
		Attributes: apipb.ToStruct(map[string]interface{}{
			"name": "coleman",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !val {
		t.Fatal("failure")
	}
}
