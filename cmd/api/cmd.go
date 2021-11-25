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
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/pubsub"
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
	ctx := logger.Inject(context.Background(), logger.Tracer{
		RemoteAddr: "system",
		AppTraceID: uuid.NewV4().String(),
	})

	// ** load config file
	configVal := &config.Config{}
	zapLog, err := config.Setup(c.configFile, configVal)
	if err != nil {
		log.Fatalf("error load config: %s", err)
		return ExitErr
	}

	// ** set global logger
	logger.SetGlobalLogger(logger.NewZap(zapLog))

	zapLog.Info("~ logger already prepared")
	logger.Info(ctx, "~ setup container")
	defaultContainer, err := container.Setup(ctx, configVal)
	if err != nil {
		logger.Error(ctx, "~ error setup container", logger.KV("error", err))
		return ExitErr
	}

	defer func() {
		logger.Info(ctx, "~ closing container")
		if _err := defaultContainer.Close(); _err != nil {
			logger.Error(ctx, "~ error close container", logger.KV("error", _err))
		}
	}()

	// ** START DEPENDENCIES
	logger.Info(ctx, "~ starting up dependencies")
	logger.Info(ctx, "~~ preparing app repo")
	appRepo, err := defaultContainer.AppRepo()
	if err != nil {
		logger.Error(ctx, "~~ error prepare app repo", logger.KV("error", err))
		return ExitErr
	}

	logger.Debug(ctx, "~~ preparing fcm repo")
	fcmRepo, err := defaultContainer.FCMRepo()
	if err != nil {
		logger.Error(ctx, "~~ error fcm repo", logger.KV("error", err))
		return ExitErr
	}

	uuidFunc := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2021, 6, 28, 00, 00, 00, 00, time.UTC),
	})

	// ** PREPARE CLIENTS
	logger.Info(ctx, "~~ prepare fcm client")
	fcmClient, err := fcm.NewClient()
	if err != nil {
		logger.Error(ctx, "~~ fcm client error", logger.KV("error", err))
		return ExitErr
	}

	// ** START SERVICES
	logger.Info(ctx, "~ setting up services")
	logger.Info(ctx, "~~ app service")
	appService, err := appservice.New(appservice.DefaultServiceConfig{
		AppRepo: appRepo,
	})
	if err != nil {
		logger.Error(ctx, "~~ setting up app service error", logger.KV("error", err))
		return ExitErr
	}

	logger.Info(ctx, "~~ FCM service")
	fcmService, err := fcmservice.New(fcmservice.DefaultServiceConfig{
		FCMRepo:    fcmRepo,
		AppService: appService,
	})
	if err != nil {
		logger.Error(ctx, "~~ setting up FCM service error", logger.KV("error", err))
		return ExitErr
	}

	logger.Info(ctx, "~~ preparing message service processor")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		FCMService: fcmService,
		FCMClient:  fcmClient,
		RESTClient: resty.New(),
	})
	if err != nil {
		logger.Error(ctx, "~~ setting up message service processor error", logger.KV("error", err))
		return ExitErr
	}

	// default using sync mode unless queue enabled
	var msgServiceDispatcher msgservice.Service = msgServiceProcessor

	if !configVal.MsgService.QueueDisable {
		logger.Debug(ctx, "preparing message service dispatcher",
			logger.KV("queueType", configVal.MsgService.QueueType),
		)
		var pubSubMsgService pubsub.IPublisher

		switch configVal.MsgService.QueueType {
		case "redis":
			// pubsub can only support redis with single architecture
			redisConn, err := defaultContainer.GetRedis().GetSingle(configVal.MsgService.QueueIdentifier)
			if err != nil {
				err = fmt.Errorf("message service with type redis %s: get connection error: %w",
					configVal.MsgService.QueueIdentifier, err)

				logger.Error(ctx, "error preparing message service dispatcher",
					logger.KV("error", err),
				)
				return ExitErr
			}

			pubSubMsgService, err = pubsub.NewRedis(pubsub.RedisConfig{
				Context:       ctx,
				QueueName:     "producer",
				CleanUpTicker: time.Hour,
				Concurrency:   configVal.Worker.Num,
				RedisClient:   redisConn,
			})

			if err != nil {
				logger.Error(ctx, "error preparing message service dispatcher",
					logger.KV("error", err),
				)
				return ExitErr
			}

			defer func() {
				if _err := pubSubMsgService.Shutdown(ctx); _err != nil {
					logger.Error(ctx, "error shutdown pubsub with redis", logger.KV("error", _err))
				}
			}()

		default:
			err = fmt.Errorf("pubsub with type %s is unknown", configVal.MsgService.QueueType)

			logger.Error(ctx, "error preparing message service dispatcher",
				logger.KV("error", err),
			)
			return ExitErr
		}

		// ** prepare message dispatcher
		msgServiceDispatcher, err = msgservice.NewDispatcher(msgservice.DispatcherConfig{
			Publisher: pubSubMsgService,
		})
		if err != nil {
			logger.Error(ctx, "error preparing message service dispatcher",
				logger.KV("error", err),
			)
			return ExitErr
		}
	}

	// ** HTTP TRANSPORT
	serverConfig := HTTPServer.Config{
		AppServiceName:    c.appName,
		AppVersion:        c.appVersion,
		DebugError:        true,
		UID:               uuidFunc,
		AppService:        appService,
		FCMService:        fcmService,
		MessageProcessor:  msgServiceProcessor,
		MessageDispatcher: msgServiceDispatcher,
	}

	logger.Info(ctx, "~ prepare http transport")
	server, err := HTTPServer.NewHTTPTransport(serverConfig)
	if err != nil {
		logger.Error(ctx, "~ prepare http transport error", logger.KV("error", err))
		return ExitErr
	}

	httpPort := fmt.Sprintf(":%d", configVal.Transport.HTTP.Port)
	logger.Debug(ctx, fmt.Sprintf("~ http transport is up on port %s", httpPort))

	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: server.Server(),
	}

	var apiErrChan = make(chan error, 1)
	go func() {
		apiErrChan <- httpServer.ListenAndServe()
	}()

	// ** listen for sigterm signal
	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		logger.Info(ctx, "exiting http server")
		if _err := httpServer.Shutdown(ctx); _err != nil {
			logger.Error(ctx, "error shutdown", logger.KV("error", _err))
		}

	case err := <-apiErrChan:
		if err != nil {
			logger.Info(ctx, "error HTTP API", logger.KV("error", err))
		}
	}

	return ExitSuccess
}

func (c *Cmd) Synopsis() string {
	return `API will start server using HTTP or gRPC`
}
