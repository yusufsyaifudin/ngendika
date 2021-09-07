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

type AccessLogData struct {
	Path        string      `json:"path,omitempty"`
	ReqBody     interface{} `json:"req_body,omitempty"`
	RespBody    interface{} `json:"resp_body,omitempty"`
	Error       string      `json:"error,omitempty"`
	ElapsedTime int64       `json:"elapsed_time,omitempty"`
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
