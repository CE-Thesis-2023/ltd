package logger

func Infof(format string, args ...interface{}) {
	Sugar().Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Sugar().Debugf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Sugar().Warnf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Sugar().Fatalf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Sugar().Errorf(format, args...)
}

func Info(args ...interface{}) {
	Sugar().Info(args...)
}

func Debug(args ...interface{}) {
	Sugar().Debug(args...)
}

func Warn(args ...interface{}) {
	Sugar().Warn(args...)
}

func Fatal(args ...interface{}) {
	Sugar().Fatal(args...)
}

func Error(args ...interface{}) {
	Sugar().Error(args...)
}

func Infoln(args ...interface{}) {
	Sugar().Infoln(args...)
}

func Debugln(args ...interface{}) {
	Sugar().Debugln(args...)
}

func Warnln(args ...interface{}) {
	Sugar().Warnln(args...)
}

func Fatalln(args ...interface{}) {
	Sugar().Fatalln(args...)
}

func Errorln(args ...interface{}) {
	Sugar().Errorln(args...)
}
