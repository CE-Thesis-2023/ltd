package service

import (
	"sync"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
)

var once sync.Once

var (
	sms                 *StreamManagementService
	commandService      *CommandService
	controlPlaneService *ControlPlaneService
)

func Init() {
	once.Do(func() {
		commandService = NewCommandService()
		controlPlaneService = NewControlPlaneService(&configs.Get().DeviceInfo)
	})
}

func GetCommandService() *CommandService {
	return commandService
}

func GetStreamManagementService() *StreamManagementService {
	return sms
}

func GetControlPlaneService() *ControlPlaneService {
	return controlPlaneService
}

func Shutdown() {
	sms.
		MediaService().
		Shutdown()
}
