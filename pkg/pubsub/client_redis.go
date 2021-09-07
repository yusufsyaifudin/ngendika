package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
)

type RedisConfig struct {
	Concurrency int `yaml:"concurrency" validate:"required"`

	Mode       string   `yaml:"mode" validate:"required,oneof=single sentinel cluster"`
	Address    []string `yaml:"address" validate:"required,unique,dive,required"`
	Username   string   `yaml:"username" validate:"-"`
	Password   string   `yaml:"password" validate:"-"`
	DB         int      `yaml:"db" validate:"-"`
	MasterName string   `yaml:"master_name" validate:"required_if=Mode sentinel"`
}

type Redis struct {
	queueTypeName string
	publisher     *asynq.Client
	subscriber    *asynq.Server
}

var _ IPublisher = (*Redis)(nil)
var _ ISubscriber = (*Redis)(nil)

func NewRedis(conf RedisConfig) (*Redis, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	client := RedisClusterClientOpt{
		Mode:       conf.Mode,
		Address:    conf.Address,
		Username:   conf.Username,
		Password:   conf.Password,
		DB:         conf.DB,
		MasterName: conf.MasterName,
	}

	publisher := asynq.NewClient(client)
	subscriber := asynq.NewServer(client, asynq.Config{
		// Specify how many concurrent workers to use
		Concurrency:    conf.Concurrency,
		RetryDelayFunc: nil,
		// Optionally specify multiple queues with different priority.
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})

	return &Redis{
		queueTypeName: "ngendika_queue::", // MUST BE STATIC, DON'T CHANGE AT RUN TIME
		publisher:     publisher,
		subscriber:    subscriber,
	}, nil
}

func (r *Redis) Publish(ctx context.Context, msg *Message) (err error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	task := asynq.NewTask(r.queueTypeName, payload)
	_, err = r.publisher.Enqueue(task)
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Subscribe(parentCtx context.Context, handler SubscribeHandler) {
	mux := asynq.NewServeMux()
	mux.HandleFunc(r.queueTypeName, func(ctx context.Context, task *asynq.Task) error {
		var msg *Message
		if err := json.Unmarshal(task.Payload(), &msg); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		msg.LoggableID = task.Type()

		err := handler(ctx, msg)
		if errors.Is(err, ErrTriggerDoNackMessage) {
			// push back the message to queue if this is not the right handler
			// TODO: this will cause message looping if there is no worker running to handle the message
			//  solution: add register type for each handle
			return r.Publish(ctx, msg)
		}

		return err
	})

	_ = r.subscriber.Run(mux) // TODO handle error
}

func (r *Redis) Shutdown(ctx context.Context) (err error) {
	r.subscriber.Stop()
	r.subscriber.Shutdown()
	return r.publisher.Close()
}