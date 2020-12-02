package stack_test

import (
	stack2 "github.com/autom8ter/graphik/generic/stack"
	"testing"
)

func Test(t *testing.T) {
	stck := stack2.New()
	stck.Push("hello")
	if stck.Len() != 1 {
		t.Fatal("expected length of one")
	}
	val := stck.Pop()
	if val.(string) != "hello" {
		t.Fatal("expected hello")
	}
}
