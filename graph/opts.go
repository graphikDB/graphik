package graph

import "time"

type Option struct {
	updatedAt time.Time
}

type Opt func(o *Option)

func WithUpdatedAt(time time.Time) Opt {
	return func(o *Option) {
		o.updatedAt = time
	}
}
