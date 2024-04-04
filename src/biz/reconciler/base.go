package reconciler

import "context"

type BaseReconciler interface {
	Run(ctx context.Context)
}
