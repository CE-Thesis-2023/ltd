package reconciler

import (
	"sync"

	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
)

var once sync.Once

var (
	reconciler *Reconciler
)

func Init() {
	once.Do(func() {
		reconciler = NewReconciler(
			service.GetControlPlaneService(),
			&configs.Get().
				DeviceInfo,
			service.GetCommandService(),
		)
	})
}

func GetReconciler() *Reconciler {
	return reconciler
}
