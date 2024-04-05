package service

import (
	"sync"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
)

var once sync.Once

var (
	sms                 MediaServiceInterface
	commandService      *CommandService
	controlPlaneService *ControlPlaneService
)

func Init() {
	once.Do(func() {
		commandService = NewCommandService()
		controlPlaneService = NewControlPlaneService(&configs.Get().DeviceInfo)
		sms = NewMediaService()
	})
}

func GetCommandService() *CommandService {
	return commandService
}

func GetControlPlaneService() *ControlPlaneService {
	return controlPlaneService
}

func GetMediaService() MediaServiceInterface {
	return sms
}

func Shutdown() {
	sms.
		Shutdown()
}
