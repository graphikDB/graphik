package graph

import "github.com/autom8ter/machine"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	machine *machine.Machine
}

func NewResolver(machine *machine.Machine) *Resolver {
	return &Resolver{machine: machine}
}
