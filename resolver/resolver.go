package resolver

import (
	"github.com/autom8ter/graphik/runtime"
	"github.com/autom8ter/machine"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	machine *machine.Machine
	store   *runtime.Store
}

func NewResolver(machine *machine.Machine, store *runtime.Store) *Resolver {
	return &Resolver{
		machine: machine,
		store:   store,
	}
}

func (r *Resolver) Close() error {
	return r.store.Close()
}
