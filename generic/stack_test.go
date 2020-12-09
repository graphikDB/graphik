package generic_test

import (
	"github.com/graphikDB/graphik/generic"
	"testing"
)

func Test(t *testing.T) {
	stck := generic.NewStack()
	stck.Push("hello")
	if stck.Len() != 1 {
		t.Fatal("expected length of one")
	}
	val := stck.Pop()
	if val.(string) != "hello" {
		t.Fatal("expected hello")
	}
}
