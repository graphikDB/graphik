package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
)

var logger *zap.Logger

func Logger(withFields ...zap.Field) *zap.Logger {
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

func Info(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	Logger().Info(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	Logger().Fatal(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	Logger().Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	Logger().Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	Logger().Error(msg, fields...)
}
