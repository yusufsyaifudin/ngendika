package tracer

type Data struct {
	RemoteAddr string `tracer:"remote_addr"`
	TraceID    string `tracer:"trace_id"`
}
