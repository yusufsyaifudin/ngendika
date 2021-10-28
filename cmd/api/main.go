package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/satori/uuid"
	"github.com/sony/sonyflake"
	"github.com/spf13/cobra"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/container"
	HTTPServer "github.com/yusufsyaifudin/ngendika/internal/http"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/pubsub"
	"go.uber.org/zap"
	_ "gocloud.dev/pubsub/mempubsub"
)

// Execute will create new cobra command with the config can be loaded from (in order): file, env, flag.
// Don't move adminCmd to global variable as it will easier to understand if we only use private variable as possible.
func Execute() *cobra.Command {
	var apiCmd = &cobra.Command{
		Use:   "api",
		Short: "API Execute",
		Long:  "API Execute will provides...",
		RunE:  Handler,
	}

	// Register flag.FlagSet into "admin" command.
	// aconfig will parse "flag" tag and register it into flag set
	// loader := aconfig.LoaderFor(new(config.cfg), aconfig.cfg{})
	// flagSet := loader.Flags()
	// apiCmd.PersistentFlags().AddGoFlagSet(flagSet)

	return apiCmd
}

// Handler will prepare all dependencies and then run actual HTTP or gRPC server when done.
func Handler(cmd *cobra.Command, args []string) error {
	ctx := logger.Inject(cmd.Context(), logger.Tracer{
		RemoteAddr: "system",
		AppTraceID: uuid.NewV4().String(),
	})

	conf := config.Config{}
	zapLog, err := config.Setup(cmd, args, &conf)
	if err != nil {
		return err
	}

	// set global logger
	logger.SetGlobalLogger(logger.NewZap(zapLog))

	zapLog.Debug("~~ prepare dependencies")
	defaultContainer, err := container.Setup(ctx, conf)
	if err != nil {
		err = fmt.Errorf("error setup dependencies: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	defer func() {
		zapLog.Debug("~~ closing dependencies")
		if _err := defaultContainer.Close(); _err != nil {
			zapLog.Error("~~ error close dependencies", zap.Error(_err))
		}
	}()

	zapLog.Debug("~ setting up services")
	zapLog.Debug("~~ app service")
	appService, err := appservice.New(appservice.Config{
		AppRepo:              defaultContainer.AppRepo(),
		FCMServerKeyRepo:     defaultContainer.FCMServerKeyRepo(),
		FCMServiceAccKeyRepo: defaultContainer.FCMServiceAccountKeyRepo(),
	})
	if err != nil {
		err = fmt.Errorf("~~ setting up app service error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	zapLog.Debug("~~ fcm client")
	fcmClient, err := fcm.NewClient()
	if err != nil {
		err = fmt.Errorf("~~ fcm client error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	zapLog.Debug("~~ setting up redis pubsub...")
	pubSubRedis := make(map[string]pubsub.IPublisher)
	for name, conn := range conf.Redis {
		ps, err := pubsub.NewRedis(pubsub.RedisConfig{
			Concurrency: conf.Worker.Num,
			Mode:        conn.Mode,
			Address:     conn.Address,
			Username:    conn.Username,
			Password:    conn.Password,
			DB:          conn.DB,
			MasterName:  conn.MasterName,
		})

		if err != nil {
			err = fmt.Errorf("connect redis %s for pubsub error: %w", name, err)
			return err
		}

		pubSubRedis[name] = ps
	}

	defer func() {
		for s, publisher := range pubSubRedis {
			if _err := publisher.Shutdown(context.Background()); _err != nil {
				zapLog.Error("error shutdown redis pubsub", zap.String("name", s), zap.Error(_err))
			}
		}
	}()

	uuidFunc := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2021, 6, 28, 00, 00, 00, 00, time.UTC),
	})

	zapLog.Debug("~~ preparing message service processor")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		AppRepo:              defaultContainer.AppRepo(),
		FCMServerKeyRepo:     defaultContainer.FCMServerKeyRepo(),
		FCMServiceAccKeyRepo: defaultContainer.FCMServiceAccountKeyRepo(),
		FCMClient:            fcmClient,
		RESTClient:           resty.New(),
	})
	if err != nil {
		err = fmt.Errorf("~~ setting up message service processor error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	zapLog.Debug("~~ preparing message service dispatcher")
	var msgServiceDispatcher msgservice.Service = msgServiceProcessor // default using sync mode unless queue enabled

	if !conf.MsgService.QueueDisable {
		var pubSubMsgService pubsub.IPublisher
		var ok bool
		switch conf.MsgService.QueueType {
		case "redis":
			pubSubMsgService, ok = pubSubRedis[conf.MsgService.QueueIdentifier]
			if !ok {
				err = fmt.Errorf("pubsub with type redis: %s for message service is not found", conf.MsgService.QueueIdentifier)
				return err
			}

		default:
			err = fmt.Errorf("pubsub with type %s is unknown", conf.MsgService.QueueType)
			return err
		}

		msgServiceDispatcher, err = msgservice.NewDispatcher(msgservice.DispatcherConfig{
			Publisher: pubSubMsgService,
		})
		if err != nil {
			return err
		}
	}

	zapLog.Debug("~ prepare transport")
	zapLog.Debug("~~ http transport")
	adminSrvConf := HTTPServer.Config{
		DebugError:        true,
		UID:               uuidFunc,
		Log:               logger.NewZap(zapLog),
		AppService:        appService,
		MessageProcessor:  msgServiceProcessor,
		MessageDispatcher: msgServiceDispatcher,
	}

	adminSrv, err := HTTPServer.NewHTTPTransport(adminSrvConf)
	if err != nil {
		err = fmt.Errorf("prepare admin server error: %w", err)
		zapLog.Error("~~ http transport error", zap.Error(err))
		return err
	}

	httpPort := fmt.Sprintf(":%d", conf.Transport.HTTP.Port)
	zapLog.Debug(fmt.Sprintf("~~ http transport is up on port %s", httpPort))
	return http.ListenAndServe(httpPort, adminSrv.Server())
}
