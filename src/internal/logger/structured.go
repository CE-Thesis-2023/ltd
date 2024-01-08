package logger

import (
	"github.com/bytedance/sonic"
	"go.uber.org/zap"
)

func SInfo(msg string, args ...zap.Field) {
	Logger().Info(msg, args...)
}

func SDebug(msg string, args ...zap.Field) {
	Logger().Debug(msg, args...)
}

func SError(msg string, args ...zap.Field) {
	Logger().Error(msg, args...)
}

func SWarn(msg string, args ...zap.Field) {
	Logger().Warn(msg, args...)
}

func SFatal(msg string, args ...zap.Field) {
	Logger().Fatal(msg, args...)
}

func Json(key string, val interface{}) zap.Field {
	resBytes, _ := sonic.Marshal(val)
	return zap.String(key, string(resBytes))
}
