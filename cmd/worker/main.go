package worker

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/satori/uuid"
	"github.com/segmentio/encoding/json"
	"github.com/spf13/cobra"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/container"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/pubsub"
	"go.uber.org/zap"
)

func Execute() *cobra.Command {
	var apiCmd = &cobra.Command{
		Use:   "worker",
		Short: "API Execute",
		Long:  "API Execute will provides...",
		RunE:  Handler,
	}

	return apiCmd
}

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

	zapLog.Debug("~~ fcm client")
	fcmClient, err := fcm.NewClient()
	if err != nil {
		err = fmt.Errorf("~~ fcm client error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	// connect to all redis config
	zapLog.Debug("~~ setting up redis pubsub...")
	pubSubRedis := make(map[string]pubsub.ISubscriber)
	for name, conn := range conf.Redis {
		subs, err := pubsub.NewRedis(pubsub.RedisConfig{
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

		pubSubRedis[name] = subs
	}

	defer func() {
		for s, publisher := range pubSubRedis {
			if _err := publisher.Shutdown(context.Background()); _err != nil {
				zapLog.Error("error shutdown redis pubsub", zap.String("name", s), zap.Error(_err))
			}
		}
	}()

	zapLog.Debug("~~ preparing message service processor")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		AppRepo:              defaultContainer.AppRepo(),
		FCMServerKeyRepo:     defaultContainer.FCMServerKeyRepo(),
		FCMServiceAccKeyRepo: defaultContainer.FCMServiceAccountKeyRepo(),
		FCMClient:            fcmClient,
		RESTClient:           resty.New().SetDebug(true),
	})
	if err != nil {
		err = fmt.Errorf("~~ setting up message service processor error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	for name, subs := range pubSubRedis {
		if conf.Worker.QueueType != "redis" {
			continue
		}

		if name != conf.Worker.QueueIdentifier {
			continue
		}

		// only subscribe when queue type and identifier is match
		subs.Subscribe(context.Background(), func(ctx context.Context, msg *pubsub.Message) error {
			var task *msgservice.Task
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				return err
			}

			_, err = msgServiceProcessor.Process(ctx, task)
			return err
		})
	}

	return nil
}
