package graph

import (
	"context"
	apipb "github.com/autom8ter/graphik/api"
	"github.com/autom8ter/graphik/vm"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/pkg/errors"
)

type triggerClient struct {
	apipb.TriggerServiceClient
	matchers []cel.Program
}

func (t *triggerClient) shouldTrigger(m apipb.Mapper) (bool, error) {
	return vm.Eval(t.matchers, m)
}

func (t *triggerClient) refresh(ctx context.Context) error {
	matchers, err := t.Filter(ctx, &empty.Empty{})
	if err != nil {
		return err
	}
	programs, err := vm.Programs(matchers.GetExpressions())
	if err != nil {
		return errors.Wrap(err, "failed to compile trigger matchers")
	}
	t.matchers = programs
	return nil
}
