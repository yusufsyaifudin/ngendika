package worker

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/cli"
	"github.com/satori/uuid"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/container"
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
	configFile string
}

func NewCmd() func() (cli.Command, error) {
	return func() (cli.Command, error) {
		cmd := &Cmd{}
		err := cmd.init()
		return cmd, err
	}
}

var _ cli.Command = (*Cmd)(nil)
var _ cli.CommandFactory = NewCmd()

func (c *Cmd) init() error {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.configFile, "config", "config.yml",
		"Config file to load")
	return nil
}

func (c *Cmd) Help() string {
	return `Worker`
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

	// ** PREPARE CLIENTS
	logger.Info(ctx, "~~ prepare fcm client")
	fcmClient, err := fcm.NewClient()
	if err != nil {
		logger.Error(ctx, "~~ fcm client error", logger.KV("error", err))
		return ExitErr
	}

	// ** START SERVICES
	logger.Info(ctx, "~ setting up services")

	logger.Info(ctx, "~~ preparing message service processor")
	msgServiceProcessor, err := msgservice.NewProcessor(msgservice.ProcessorConfig{
		AppRepo:    appRepo,
		FCMRepo:    fcmRepo,
		FCMClient:  fcmClient,
		RESTClient: resty.New(),
	})
	if err != nil {
		logger.Error(ctx, "~~ setting up message service processor error", logger.KV("error", err))
		return ExitErr
	}

	logger.Debug(ctx, "preparing worker",
		logger.KV("queueType", configVal.Worker.QueueType),
	)
	var pubSubWorker pubsub.ISubscriber
	switch configVal.Worker.QueueType {
	case "redis":
		// pubsub can only support redis with single architecture
		redisConn, err := defaultContainer.GetRedis().GetSingle(configVal.Worker.QueueIdentifier)
		if err != nil {
			err = fmt.Errorf("pubsub with type redis: %s get connection error: %w", err)
			logger.Error(ctx, "worker error", logger.KV("error", err))
			return ExitErr
		}

		pubSubWorker, err = pubsub.NewRedis(pubsub.RedisConfig{
			Concurrency: configVal.Worker.Num,
			RedisClient: redisConn,
		})
	}

	if err != nil {
		err = fmt.Errorf("pubsub with type %s: %s error: %w",
			configVal.Worker.QueueType, configVal.MsgService.QueueIdentifier, err,
		)

		logger.Error(ctx, "worker error", logger.KV("error", err))
		return ExitErr
	}

	defer func() {
		if _err := pubSubWorker.Shutdown(ctx); _err != nil {
			logger.Error(ctx, "error shutdown pubsub with redis", logger.KV("error", _err))
		}
	}()

	// only subscribe when queue type and identifier is match
	pubSubWorker.Subscribe(context.Background(), func(ctx context.Context, msg *pubsub.Message) error {
		var task *msgservice.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			return err
		}

		_, err = msgServiceProcessor.Process()(ctx, task)
		return err
	})

	return ExitSuccess
}

func (c *Cmd) Synopsis() string {
	return `Worker will process the message`
}
