package pubsub

import (
	"context"
	"fmt"
)

var (
	ErrTriggerDoNackMessage = fmt.Errorf("error processing message and do nack message when possible")
)

type IPublisher interface {
	Publish(ctx context.Context, msg *Message) (err error)
	Shutdown(ctx context.Context) (err error)
}

type SubscribeHandler = func(ctx context.Context, msg *Message) error

type ISubscriber interface {
	Subscribe(ctx context.Context, handler SubscribeHandler)
	Shutdown(ctx context.Context) error
}

// ====== cfg common

type ConfigKafkaPublisher struct {
	Brokers []string `validate:"required,unique"`
	Topic   string   `validate:"required"`
}

type ConfigKafkaSubscriber struct {
	Brokers []string `validate:"required,unique"`
	Topic   string   `validate:"required"`
}
