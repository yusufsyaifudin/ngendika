package logger

import (
	"context"

	"go.uber.org/zap"
)

var globalLogger Logger

func getGlobalLogger() Logger {
	if globalLogger == nil {
		z, _ := zap.NewDevelopment()
		return NewZap(z)
	}
	return globalLogger
}

func SetGlobalLogger(log Logger) {
	globalLogger = log
}

func Debug(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...KeyValue) {
	getGlobalLogger().Error(ctx, msg, fields...)
}

func Access(ctx context.Context, data AccessLogData) {
	getGlobalLogger().Access(ctx, data)
}
