package api

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ylog"

	HTTPServer "github.com/yusufsyaifudin/ngendika/internal/http"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/cli"
	"github.com/satori/uuid"
	"github.com/sony/sonyflake"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/container"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

const (
	ExitSuccess = 0
	ExitErr     = -1
)

type Cmd struct {
	flags      *flag.FlagSet
	appName    string
	appVersion string
	configFile string
	cfg        *config.Config
	zapLog     *zap.Logger
}

func NewCmd(appName, appVersion string) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		cmd := &Cmd{
			flags:      &flag.FlagSet{},
			appName:    appName,
			appVersion: appVersion,
		}
		err := cmd.init()
		return cmd, err
	}
}

var _ cli.Command = (*Cmd)(nil)
var _ cli.CommandFactory = NewCmd("", "")

func (c *Cmd) init() error {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.configFile, "config", "config.yml",
		"Config file to load")
	c.flags.StringVar(&c.configFile, "c", "config.yml",
		"Alias for config file to load")

	// ** load config file
	zapLog, err := config.Setup(c.configFile, &c.cfg)
	if err != nil {
		err = fmt.Errorf("error load config: %w", err)
		return err
	}

	c.zapLog = zapLog

	return nil
}

func (c *Cmd) Help() string {
	return `API will start server using HTTP or gRPC`
}

func (c *Cmd) Run(args []string) int {
	err := c.flags.Parse(args)
	if err != nil {
		log.Fatalf("error parsing config argument: %s", err)
		return ExitErr
	}

	// ** define system context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	propagateData := tracer.Data{
		RemoteAddr: "system",
		TraceID:    uuid.NewV4().String(),
	}

	traceLog, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
	if err != nil {
		log.Fatalf("error prepare tracer system data: %s", err)
		return ExitErr
	}

	// inject context
	ctx = ylog.Inject(ctx, traceLog)

	// ** set global logger
	ylog.SetGlobalLogger(ylog.NewZap(c.zapLog))

	ylog.Info(ctx, "config is loaded and logger is prepared")
	ylog.Info(ctx, "container preparation: starting")
	defaultContainer, err := container.Setup(ctx, c.cfg)
	if err != nil {
		ylog.Error(ctx, "container preparation: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "container preparation: done")
	defer func() {
		ylog.Info(ctx, "trying closing container")
		if _err := defaultContainer.Close(); _err != nil {
			ylog.Error(ctx, "error close container", ylog.KV("error", _err))
		}
	}()

	// ** START DEPENDENCIES
	ylog.Info(ctx, "dependencies preparation: starting")
	ylog.Info(ctx, "app repository: starting")
	appRepo, err := defaultContainer.AppRepo()
	if err != nil {
		ylog.Error(ctx, "app repository: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "fcm repository: starting")
	fcmRepo, err := defaultContainer.FCMRepo()
	if err != nil {
		ylog.Error(ctx, "fcm repository: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "dependencies preparation: done")

	// ** UUID function
	uuidFunc := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2021, 6, 28, 00, 00, 00, 00, time.UTC),
	})

	// ** PREPARE CLIENTS
	ylog.Info(ctx, "client(s) preparation: starting")
	ylog.Info(ctx, "fcm client: starting")
	var fcmClient fcm.Client
	fcmClient, err = fcm.NewClient()
	if err != nil {
		ylog.Error(ctx, "fcm client: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "client(s) preparation: done")

	// ** START SERVICES
	ylog.Info(ctx, "services preparation: starting")
	ylog.Info(ctx, "APP service: starting")
	appService, err := appservice.New(appservice.DefaultServiceConfig{
		AppRepo: appRepo,
	})
	if err != nil {
		ylog.Error(ctx, "app service: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "FCM service: starting")
	fcmService, err := fcmservice.New(fcmservice.DefaultServiceConfig{
		FCMRepo:    fcmRepo,
		AppService: appService,
		FCMClient:  fcmClient,
	})
	if err != nil {
		ylog.Error(ctx, "FCM service: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "message service processor: starting")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		FCMService: fcmService,
		RESTClient: resty.New(),
		MaxWorker:  c.cfg.MsgService.MaxParallel,
	})
	if err != nil {
		ylog.Error(ctx, "message service processor: failed", ylog.KV("error", err))
		return ExitErr
	}

	ylog.Info(ctx, "transport preparation: starting")

	// ** HTTP TRANSPORT
	serverConfig := HTTPServer.Config{
		AppServiceName:   c.appName,
		AppVersion:       c.appVersion,
		DebugError:       true,
		UID:              uuidFunc,
		AppService:       appService,
		FCMService:       fcmService,
		MessageProcessor: msgServiceProcessor,
	}

	ylog.Info(ctx, "http transport: starting")
	server, err := HTTPServer.NewHTTPTransport(serverConfig)
	if err != nil {
		ylog.Error(ctx, "http transport: failed", ylog.KV("error", err))
		return ExitErr
	}

	httpPort := fmt.Sprintf(":%d", c.cfg.Transport.HTTP.Port)
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: server.Server(),
	}

	var apiErrChan = make(chan error, 1)
	go func() {
		ylog.Info(ctx, fmt.Sprintf("http transport: done running on port %d", c.cfg.Transport.HTTP.Port))
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

	return ExitSuccess
}

func (c *Cmd) Synopsis() string {
	return `API will start server using HTTP or gRPC`
}
