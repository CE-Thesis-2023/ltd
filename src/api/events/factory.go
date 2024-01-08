package eventsapi

import (
	"context"
	"sync"
)

var once sync.Once

var standardEventsHandler *StandardEventHandler

func Init(ctx context.Context) {
	once.Do(func() {
		standardEventsHandler = NewStandardEventHandler()
	})
}

func GetStandardEventsHandler() *StandardEventHandler {
	return standardEventsHandler
}
