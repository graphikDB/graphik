package logger

import (
	"github.com/graphikDB/graphik/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
)

type Logger struct {
	logger *zap.Logger
}

func New(debug bool, withFields ...zap.Field) *Logger {
	hst, _ := os.Hostname()
	withFields = append(withFields, zap.String("host", hst))
	withFields = append(withFields, zap.String("service", "graphik"))
	withFields = append(withFields, zap.String("version", version.Version))

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
	core := zapcore.NewCore(jsonEncoder, os.Stdout, zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if lvl == zap.DebugLevel {
			return debug
		}
		return true
	}))
	return &Logger{
		logger: zap.New(core).With(withFields...),
	}
}

func appendFields(fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.Int("goroutines", runtime.NumGoroutine()))
	return fields
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, appendFields(fields...)...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, appendFields(fields...)...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, appendFields(fields...)...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, appendFields(fields...)...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, appendFields(fields...)...)
}

func (l *Logger) Zap() *zap.Logger {
	return l.logger
}
