package custinflux

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type RegisterFunc func(ctx context.Context, c influxdb2.Client)

type Options struct {
	registerFunc RegisterFunc
}

type Optioner func(o *Options)

func WithRegisterFunc(f RegisterFunc) Optioner {
	return func(o *Options) {
		o.registerFunc = f
	}
}
