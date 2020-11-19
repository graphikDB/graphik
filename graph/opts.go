package graph

import "time"

type Options struct {
	updatedAt time.Time
	createdAt time.Time
}

type Opt func(o *Options)

func WithUpdatedAt(time time.Time) Opt {
	return func(o *Options) {
		o.updatedAt = time
	}
}

func WithCreatedAt(time time.Time) Opt {
	return func(o *Options) {
		o.createdAt = time
	}
}
