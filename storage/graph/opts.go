package graph

import (
	apipb "github.com/autom8ter/graphik/api"
)

type Options struct {
	metadata *apipb.Metadata
}

type Opt func(o *Options)

func WithMetadata(metadata *apipb.Metadata) Opt {
	return func(o *Options) {
		o.metadata = metadata
	}
}
