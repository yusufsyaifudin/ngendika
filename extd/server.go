package extd

import (
	"context"
	"fmt"
	"github.com/satori/uuid"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/backend/befcm"
	"github.com/yusufsyaifudin/ngendika/container"
	"github.com/yusufsyaifudin/ngendika/pkg/httplog"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/transport/restapi"
	"github.com/yusufsyaifudin/ylog"
	jaegerPropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// RunServer located in extd (extended) to add capability extends backend if you want to create custom backend.
func RunServer(ctx context.Context, cfg container.Config) (err error) {

	if ctx == nil {
		ctx = context.TODO()
	}

	ctx = setupLog(ctx)

	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")),
	)
	if err != nil {
		ylog.Error(ctx, "cannot setup jaeger exporter", ylog.KV("error", err))
		return
	}

	tracer.InitTraceProvider(exp)

	// register ot propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		&ot.OT{},
		&jaegerPropagator.Jaeger{},
	))

	// ** setup repositories
	ylog.Info(ctx, "container preparation: starting")
	var repositories container.Repositories
	repositories, err = container.SetupRepositories(cfg.DatabaseResources)
	defer func() {
		ylog.Info(ctx, "closing container: starting")
		if repositories == nil {
			ylog.Info(ctx, "closing container: no need to close")
			return
		}

		if _err := repositories.Close(); _err != nil {
			ylog.Error(ctx, "closing container: failed", ylog.KV("error", _err))
		}

		ylog.Info(ctx, "closing container: done")
	}()

	if err != nil {
		ylog.Error(ctx, "container preparation: failed", ylog.KV("error", err))
		return
	}

	ylog.Info(ctx, "container preparation: done")

	// ** START SERVICES using configured repositories
	ylog.Info(ctx, "services preparation: starting")
	services, err := container.SetupServices(cfg.Services, repositories)
	if err != nil {
		ylog.Error(ctx, "service preparation: failed", ylog.KV("error", err))
		return
	}

	// ** HTTP TRANSPORT
	ylog.Info(ctx, "transport preparation: starting")
	serverConfig := restapi.Config{
		AppServiceName: "app name",
		AppVersion:     "1.0.0",
		AppService:     services.App(),
		PNPService:     services.PushNotificationProvider(),
		MsgService:     services.Message(),
	}

	ylog.Info(ctx, "http transport: starting")
	server, err := restapi.NewHTTPTransport(serverConfig)
	if err != nil {
		ylog.Error(ctx, "http transport: failed", ylog.KV("error", err))
		return
	}

	httpPort := fmt.Sprintf(":%d", cfg.Transport.HTTP.Port)
	h2s := &http2.Server{}
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: h2c.NewHandler(server.Server(), h2s), // HTTP/2 Cleartext handler
	}

	var apiErrChan = make(chan error, 1)
	go func() {
		ylog.Info(ctx, fmt.Sprintf("http transport: done running on port %d", cfg.Transport.HTTP.Port))
		apiErrChan <- httpServer.ListenAndServe()
	}()

	ylog.Info(ctx, "system: up and running...")

	// ** listen for sigterm signal
	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		ylog.Info(ctx, "system: exiting...")
		ylog.Info(ctx, "http transport: exiting...")
		if _err := httpServer.Shutdown(ctx); _err != nil {
			ylog.Error(ctx, "http transport: ", ylog.KV("error", _err))
		}

	case err := <-apiErrChan:
		if err != nil {
			ylog.Info(ctx, "http transport: error", ylog.KV("error", err))
		}
	}

	return
}

func setupLog(ctx context.Context) context.Context {

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			MessageKey:     "msg",
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			LineEnding:     zapcore.DefaultLineEnding,
			LevelKey:       "level",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
		}),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), // pipe to multiple writer
		zapcore.DebugLevel,
	)

	zapLog := zap.New(core)

	propagateData := tracer.LogData{
		RemoteAddr: "system",
		TraceID:    uuid.NewV4().String(),
	}

	traceLog, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
	if err != nil {
		log.Fatalf("error prepare tracer system data: %s", err)
		return ctx
	}

	// inject context
	ctx = ylog.Inject(ctx, traceLog)

	// ** set global logger
	ylog.SetGlobalLogger(ylog.NewZap(zapLog))

	return ctx
}

func RegisterDefaultBackends(ctx context.Context) (err error) {
	ylog.Info(ctx, "httplog for outgoing")
	httpLogOut, err := httplog.New()
	if err != nil {
		err = fmt.Errorf("http log preparation failed: %w", err)
		return
	}

	err = backend.Register("noop", backend.NewNoopSender())
	if err != nil {
		err = fmt.Errorf("register backend noop failed: %w", err)
		return
	}

	beFcm, err := befcm.NewBE(httpLogOut)
	if err != nil {
		err = fmt.Errorf("be fcm failed: %w", err)
		return
	}

	err = backend.Register("fcm", beFcm)
	if err != nil {
		err = fmt.Errorf("register backend fcm failed: %w", err)
		return
	}

	return
}
