package logger

import (
	"context"

	"go.uber.org/zap"
)

const (
	TypeAccessLog = "access_log"
	TypeSys       = "sys"
)

type Zap struct {
	writer *zap.Logger
}

func NewZap(zapLogger *zap.Logger) *Zap {
	return &Zap{writer: zapLogger}
}

func (z *Zap) Debug(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Debug(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Info(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Info(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Warn(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Warn(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Error(ctx context.Context, msg string, fields ...KeyValue) {
	z.writer.Error(msg, localFieldZapFields(ctx, TypeSys, fields)...)
}

func (z *Zap) Access(ctx context.Context, data AccessLogData) {
	z.writer.Info(TypeAccessLog, localFieldZapFields(ctx, TypeAccessLog, []KeyValue{KV("data", data)})...)
}

func localFieldZapFields(ctx context.Context, tag string, fields []KeyValue) []zap.Field {
	zapFields := make([]zap.Field, 0)
	zapFields = append(zapFields, zap.String("tag", tag))

	data, ok := Extract(ctx)
	if ok {
		zapFields = append(zapFields, zap.Any("tracer", data))
	}

	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}

	return zapFields
}

var _ Logger = (*Zap)(nil)
