package generic_test

import (
	"context"
	"github.com/autom8ter/machine"
	"github.com/graphikDB/graphik/generic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cache := generic.NewCache(machine.New(context.Background()), 1*time.Minute)
	cache.Set("key", "value", 0)
	if cache.Len() != 1 {
		t.Fatal("expected 1 key")
	}
}
