package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
)

type RedisConfig struct {
	Context       context.Context `validate:"required"`
	QueueName     string          `validate:"required"`
	CleanUpTicker time.Duration   `validate:"required"`
	Concurrency   int             `validate:"required"`
	RedisClient   *redis.Client   ` validate:"required,structonly"`
}

type Redis struct {
	conf  RedisConfig
	queue rmq.Queue
}

var _ IPublisher = (*Redis)(nil)
var _ ISubscriber = (*Redis)(nil)

func NewRedis(conf RedisConfig) (*Redis, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	var errChan = make(chan error)
	conn, err := rmq.OpenConnectionWithRedisClient("producer", conf.RedisClient, errChan)
	if err != nil {
		return nil, err
	}

	queue, err := conn.OpenQueue(conf.QueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		cleaner := rmq.NewCleaner(conn)
		for range time.Tick(conf.CleanUpTicker) {
			returned, err := cleaner.Clean()
			if err != nil {
				logger.Error(conf.Context, fmt.Sprintf("failed to clean redis queue '%s'", conf.QueueName),
					logger.KV("error", err),
				)
				continue
			}
			logger.Debug(conf.Context, fmt.Sprintf("cleaned %d from redis queue '%s'", returned, conf.QueueName))
		}
	}()

	return &Redis{
		conf:  conf,
		queue: queue,
	}, nil
}

func (r *Redis) Publish(ctx context.Context, msg *Message) (err error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = r.queue.PublishBytes(payload)
	return
}

func (r *Redis) Subscribe(parentCtx context.Context, handler SubscribeHandler) {
	const (
		prefetchLimit = 1000
		pollDuration  = 100 * time.Millisecond
	)

	if err := r.queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		logger.Error(r.conf.Context, "error starting the consumer", logger.KV("error", err))
		return
	}

	for i := 0; i < r.conf.Concurrency; i++ {
		name := fmt.Sprintf("consumer %d", i)

		_, err := r.queue.AddConsumerFunc(name, func(delivery rmq.Delivery) {
			err := handler(parentCtx, &Message{
				LoggableID: name,
				Body:       []byte(delivery.Payload()),
			})
			if errors.Is(err, ErrTriggerDoNackMessage) {
				logger.Error(r.conf.Context, "push back message to dead letter queue", logger.KV("error", err))

				err = delivery.Push() // push back to rejected message
				if err != nil {
					logger.Error(r.conf.Context, "error push into dead letter queue", logger.KV("error", err))
					return
				}

				err = nil // discard error
			}

			if err != nil {
				logger.Error(r.conf.Context, "error handle message", logger.KV("error", err))
				return
			}

		})

		if err != nil {
			logger.Error(r.conf.Context, "error adding new consumer", logger.KV("error", err))
			return
		}
	}
}

func (r *Redis) Shutdown(ctx context.Context) (err error) {
	r.queue.StopConsuming()
	return
}
