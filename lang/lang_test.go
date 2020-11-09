package lang_test

import (
	"github.com/autom8ter/graphik/lang"
	"testing"
)

func Test(t *testing.T) {
	val, err := lang.BooleanExpression([]string{`_type.startsWith("user")`}, lang.NewValues(map[string]interface{}{
		"_type": "user",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !val {
		t.Fatal("failure")
	}
}
