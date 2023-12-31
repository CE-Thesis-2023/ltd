package custcon

import (
	"context"

	"github.com/carlmjohnson/flowmatic"
)

func Do(tasks ...func() error) error {
	return flowmatic.Do(tasks...)
}

func Each[T interface{}](size int, items []T, task func(item T) error) error {
	return flowmatic.Each[T](size, items, task)
}

func DoCtx(ctx context.Context, tasks ...func(ctx context.Context) error) error {
	return flowmatic.All(ctx, tasks...)
}
