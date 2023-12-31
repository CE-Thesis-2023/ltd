package service

import "sync"

var once sync.Once

var (
	streamManagementService *StreamManagementService
	commandService          *CommandService
)

func Init() {
	once.Do(func() {
		streamManagementService = NewStreamManagementService()
		commandService = NewCommandService()
	})
}

func GetCommandService() *CommandService {
	return commandService
}

func GetStreamManagementService() *StreamManagementService {
	return streamManagementService
}
