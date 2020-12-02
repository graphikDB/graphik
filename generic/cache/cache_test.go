package cache_test

import (
	"context"
	cache2 "github.com/autom8ter/graphik/generic/cache"
	"github.com/autom8ter/machine"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cache := cache2.New(machine.New(context.Background()), 1*time.Minute)
	cache.Set("key", "value", 0)
	if cache.Len() != 1 {
		t.Fatal("expected 1 key")
	}
}
