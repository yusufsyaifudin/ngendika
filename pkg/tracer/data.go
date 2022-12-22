package tracer

type LogData struct {
	RemoteAddr string `tracer:"remote_addr"`
	TraceID    string `tracer:"trace_id"`
}
