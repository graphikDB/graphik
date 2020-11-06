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
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "ts",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    "function",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		})
		fn := zap.LevelEnablerFunc(func(zapcore.Level) bool { return true })
		core := zapcore.NewCore(jsonEncoder, os.Stdout, fn)
		logger = zap.New(core).With(withFields...)
	}
	return logger
}

func appendFields(fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	return fields
}

func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, appendFields(fields...)...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger().Fatal(msg, appendFields(fields...)...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, appendFields(fields...)...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger().Debug(msg, appendFields(fields...)...)
}

func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, appendFields(fields...)...)
}
