package logger

import (
	"github.com/autom8ter/graphik/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
)

var logger *zap.Logger

func Logger(withFields ...zap.Field) *zap.Logger {
	withFields = append(withFields, zap.String("service", "graphik"))
	withFields = append(withFields, zap.String("version", version.Version))
	if logger == nil {
		zap.NewDevelopmentConfig()
		jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		})
		core := zapcore.NewCore(jsonEncoder, os.Stdout, zap.LevelEnablerFunc(func(zapcore.Level) bool { return true }))
		logger = zap.New(core)
		if len(withFields) > 0 {
			logger = logger.With(withFields...)
		}
	}
	return logger
}

func appendFields(fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	_, file, ln, ok := runtime.Caller(1)
	if ok {
		fields = append(fields, zap.String("file", file))
		fields = append(fields, zap.Int("line", ln))
	}
}

func Info(msg string, fields ...zap.Field) {
	appendFields(fields...)
	Logger().Info(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	appendFields(fields...)
	Logger().Fatal(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	appendFields(fields...)
	Logger().Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	appendFields(fields...)
	Logger().Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	appendFields(fields...)
	Logger().Error(msg, fields...)
}
