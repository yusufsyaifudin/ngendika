package pubsub

import (
	"context"
	"fmt"

	cdkPubSub "gocloud.dev/pubsub"

	"github.com/go-playground/validator/v10"
	"gocloud.dev/pubsub/kafkapubsub"
)

type PublisherConfig struct {
	Scheme string               `validate:"required,oneof=kafka"`
	Kafka  ConfigKafkaPublisher `validate:"required_if=Scheme kafka"`
}

type Publisher struct {
	topic *cdkPubSub.Topic
}

var _ IPublisher = (*Publisher)(nil)

func (c *Publisher) Publish(ctx context.Context, msg *Message) (err error) {
	err = c.topic.Send(ctx, &cdkPubSub.Message{
		Body: msg.Body,
	})

	return
}

func (c *Publisher) Shutdown(ctx context.Context) error {
	return c.topic.Shutdown(ctx)
}

var _ IPublisher = (*Publisher)(nil)

func NewPublisher(conf PublisherConfig) (*Publisher, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	var topic *cdkPubSub.Topic
	switch conf.Scheme {
	case "kafka":
		topic, err = KafkaPubOpener(conf.Kafka)
	}

	if err != nil {
		return nil, err
	}

	return &Publisher{topic: topic}, nil
}

func KafkaPubOpener(conf ConfigKafkaPublisher) (*cdkPubSub.Topic, error) {
	kafkaConfig := kafkapubsub.MinimalConfig()
	pubTopic, err := kafkapubsub.OpenTopic(conf.Brokers, kafkaConfig, conf.Topic, nil)
	if err != nil {
		return nil, fmt.Errorf("pubsub kafka error: %w", err)
	}

	return pubTopic, nil
}
