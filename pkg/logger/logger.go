package logger

import (
	"context"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...KeyValue)
	Info(ctx context.Context, msg string, fields ...KeyValue)
	Warn(ctx context.Context, msg string, fields ...KeyValue)
	Error(ctx context.Context, msg string, fields ...KeyValue)
	Access(ctx context.Context, data AccessLogData)
}

type HTTPData struct {
	Header     map[string]string `json:"header,omitempty"`
	DataObject interface{}       `json:"data_object,omitempty"`
	DataString string            `json:"data_string,omitempty"`
}

type AccessLogData struct {
	Path        string   `json:"path,omitempty"`
	Request     HTTPData `json:"request,omitempty"`
	Response    HTTPData `json:"response,omitempty"`
	Error       string   `json:"error,omitempty"`
	ElapsedTime int64    `json:"elapsed_time,omitempty"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func KV(k string, v interface{}) KeyValue {
	return KeyValue{
		Key:   k,
		Value: v,
	}
}
