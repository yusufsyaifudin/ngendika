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

	zapLog.Debug("~ injecting dependencies")
	zapLog.Debug("~~ preparing app repo")
	appRepo, err := defaultContainer.AppRepo()
	if err != nil {
		return err
	}

	zapLog.Debug("~~ preparing fcm server key repo")
	fcmServerKeyRepo, err := defaultContainer.FCMServerKeyRepo()
	if err != nil {
		return err
	}

	zapLog.Debug("~~ preparing fcm service account key app repo")
	fcmSvcAccKeyRepo, err := defaultContainer.FCMServiceAccountKeyRepo()
	if err != nil {
		return err
	}

	zapLog.Debug("~~ fcm client")
	fcmClient, err := fcm.NewClient()
	if err != nil {
		err = fmt.Errorf("~~ fcm client error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	zapLog.Debug("~~ preparing message service processor")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		AppRepo:              appRepo,
		FCMServerKeyRepo:     fcmServerKeyRepo,
		FCMServiceAccKeyRepo: fcmSvcAccKeyRepo,
		FCMClient:            fcmClient,
		RESTClient:           resty.New().SetDebug(true),
	})
	if err != nil {
		err = fmt.Errorf("~~ setting up message service processor error: %w", err)
		zapLog.Error(err.Error())
		return err
	}

	var pubsubWorker pubsub.ISubscriber
	switch conf.Worker.QueueType {
	case "redis":
		redisConn, err := defaultContainer.GetRedisConn(conf.Worker.QueueIdentifier)
		if err != nil {
			err = fmt.Errorf("pubsub with type redis: %s get connection error: %w", err)
			return err
		}

		pubsubWorker, err = pubsub.NewRedis(pubsub.RedisConfig{
			Concurrency: conf.Worker.Num,
			RedisClient: redisConn,
		})
	}

	if err != nil {
		err = fmt.Errorf("pubsub with type %s: %s error: %w",
			conf.Worker.QueueType, conf.MsgService.QueueIdentifier, err,
		)
		return err
	}

	defer func() {
		if _err := pubsubWorker.Shutdown(ctx); _err != nil {
			logger.Error(ctx, "error shutdown pubsub with redis", logger.KV("error", _err))
		}
	}()

	// only subscribe when queue type and identifier is match
	pubsubWorker.Subscribe(context.Background(), func(ctx context.Context, msg *pubsub.Message) error {
		var task *msgservice.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			return err
		}

		_, err = msgServiceProcessor.Process(ctx, task)
		return err
	})

	return nil
}
