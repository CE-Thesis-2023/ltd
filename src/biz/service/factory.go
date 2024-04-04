package service

import (
	"sync"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
)

var once sync.Once

var (
	streamManagementService *StreamManagementService
	commandService          *CommandService
	controlPlaneService     *ControlPlaneService
)

func Init() {
	once.Do(func() {
		streamManagementService = NewStreamManagementService()
		commandService = NewCommandService()
		controlPlaneService = NewControlPlaneService(&configs.Get().DeviceInfo)
	})
}

func GetCommandService() *CommandService {
	return commandService
}

func GetStreamManagementService() *StreamManagementService {
	return streamManagementService
}

func GetControlPlaneService() *ControlPlaneService {
	return controlPlaneService
}

func Shutdown() {
	GetStreamManagementService().
		MediaService().
		Shutdown()
}
