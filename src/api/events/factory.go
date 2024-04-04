package eventsapi

import (
	"sync"
)

var once sync.Once

var standardEventsHandler *StandardEventHandler

func Init() {
	once.Do(func() {
		standardEventsHandler = NewStandardEventHandler()
	})
}

func GetStandardEventsHandler() *StandardEventHandler {
	return standardEventsHandler
}
