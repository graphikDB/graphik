package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/vm"
	"github.com/google/cel-go/cel"
)

type triggerClient struct {
	apipb.TriggerServiceClient
	matchers []cel.Program
}

func (t *triggerClient) shouldTrigger(m apipb.Mapper) (bool, error) {
	return vm.Eval(t.matchers, m)
}
