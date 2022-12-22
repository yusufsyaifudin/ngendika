package tracer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/assets"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
)

const tracerAppNme = "ngendika"

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.GetTracerProvider().Tracer(tracerAppNme).Start(ctx, spanName, opts...)
}

func InitTraceProvider(exp sdktrace.SpanExporter) {
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exp),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(assets.ServiceName),
			attribute.String("environment", "development"),
		)),
	)

	otel.SetTracerProvider(tp)
}

type MiddlewareConfig struct {
	TracerName     string                        `validate:"required"`
	ServiceName    string                        `validate:"required"`
	SkipFunc       func(r *http.Request) bool    `validate:"-"`
	TracerProvider trace.TracerProvider          `validate:"required"`
	TextPropagator propagation.TextMapPropagator `validate:"required"`
}

func Middleware(cfg MiddlewareConfig, next http.Handler) http.HandlerFunc {
	if _err := validator.New().Struct(cfg); _err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}
	}

	if cfg.SkipFunc == nil {
		cfg.SkipFunc = func(r *http.Request) bool {
			return false
		}
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		if cfg.SkipFunc(r) {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		if ctx == nil {
			ctx = context.TODO()
		}

		ctx = cfg.TextPropagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(cfg.ServiceName, r.URL.Path, r)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		spanName := r.URL.Path
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", r.Method)
		}

		//tracerProvider := otel.GetTracerProvider()
		newCtx, span := cfg.TracerProvider.Tracer(cfg.TracerName).Start(ctx, spanName, opts...)
		defer span.End()

		respRec := httptest.NewRecorder()
		r = r.WithContext(newCtx)
		next.ServeHTTP(respRec, r)

		attrs := semconv.HTTPAttributesFromHTTPStatusCode(respRec.Code)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(respRec.Code, oteltrace.SpanKindServer)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)

		for k, v := range respRec.Header() {
			w.Header()[k] = v
		}

		// inject to header response
		cfg.TextPropagator.Inject(newCtx, propagation.HeaderCarrier(w.Header()))

		w.WriteHeader(respRec.Code)
		if respRec.Body != nil {
			if _, _err := bytes.NewReader(respRec.Body.Bytes()).WriteTo(w); _err != nil {
				_err = fmt.Errorf("write response body error: %w", _err)
				span.RecordError(_err)
			}
		}

	}

	return fn
}
