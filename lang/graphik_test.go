package lang_test

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/lang"
	"testing"
)

func Test(t *testing.T) {
	val, err := lang.BooleanExpression([]string{`path.startsWith("user")`}, &apipb.Node{
		Path: "user",
		Attributes: lang.ToStruct(map[string]interface{}{
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
