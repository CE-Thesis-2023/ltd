package helper

import (
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	"github.com/CE-Thesis-2023/ltd/internal/logger"

	"go.uber.org/zap"
)

var commonEventMessage string = "events handler error"

func EventHandlerErrorHandler(err error) {
	custError, ok := err.(*custerror.CustomError)
	if ok {
		logger.SInfo(commonEventMessage,
			zap.Error(err),
			zap.Uint32("type", custError.Code))
	} else {
		logger.SInfo(commonEventMessage,
			zap.Error(err))
	}
}

func RunWithRestart(f func() error, shouldRestart func(err error) bool) {
	for {
		if err := f(); err != nil {
			logger.SDebug("RunWithRestart: task failed", zap.Error(err))
			if shouldRestart(err) {
				logger.SDebug("RunWithRestart: should restart")
				continue
			}
			logger.SDebug("RunWithRestart: should stop")
			return
		}

	}
}
