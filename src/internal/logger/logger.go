package logger

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once

func Init(ctx context.Context, options ...Optioner) {
	once.Do(func() {
		opts := Options{}
		for _, opt := range options {
			opt(&opts)
		}

		logger, err := createLogger(opts.globalConfigs)
		if err != nil {
			log.Fatal(err)
		}

		zap.ReplaceGlobals(logger)
	})
}

func Sugar() *zap.SugaredLogger {
	return zap.L().Sugar()
}

func Logger() *zap.Logger {
	return zap.L()
}

func Close() {
	Logger().Sync()
}

func createLogger(opts *configs.LoggerConfigs) (*zap.Logger, error) {
	lvl, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	logConfigs := zap.Config{
		Level:             zap.NewAtomicLevelAt(*lvl),
		DisableCaller:     true,
		DisableStacktrace: true,
		Development:       false,
		Encoding:          opts.Encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	logger, err := logConfigs.Build()
	if err != nil {
		return nil, custerror.FormatInternalError("createLogger: create logger err = %s", err)
	}

	return logger, nil
}

func parseLevel(level string) (*zapcore.Level, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, custerror.FormatInvalidArgument("parseLevel: log level invalid level = %s", level)
	}
	return &lvl, nil
}

type Options struct {
	globalConfigs *configs.LoggerConfigs
}

type Optioner func(opts *Options)

func WithGlobalConfigs(globalConfigs *configs.LoggerConfigs) Optioner {
	return func(opts *Options) {
		opts.globalConfigs = globalConfigs
	}
}
