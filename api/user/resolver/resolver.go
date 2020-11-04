package resolver

import "github.com/autom8ter/graphik/lib/runtime"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	runtime *runtime.Runtime
}

func NewResolver(store *runtime.Runtime) *Resolver {
	return &Resolver{
		runtime: store,
	}
}

func (r *Resolver) Close() error {
	return r.runtime.Close()
}
