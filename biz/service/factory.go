package service

import "sync"

var once sync.Once

var (
	streamManagementService *StreamManagementService
	commandService          CommandServiceInterface
)

func Init() {
	once.Do(func() {
		streamManagementService = NewStreamManagementService()
		commandService = NewCommandService()
	})
}

func GetCommandService() CommandServiceInterface {
	return commandService
}

func GetStreamManagementService() *StreamManagementService {
	return streamManagementService
}

func Shutdown() {
	GetStreamManagementService().
		MediaService().
		Shutdown()
}
