package pubsub

import (
	"context"
	"errors"
	"fmt"
	"log"

	cdkPubSub "gocloud.dev/pubsub"

	"github.com/go-playground/validator/v10"
	"gocloud.dev/pubsub/kafkapubsub"
)

type SubscriberConfig struct {
	MaxHandler int                   `validate:"required,min=1"`
	Scheme     string                `validate:"required,oneof=kafka"`
	Kafka      ConfigKafkaSubscriber `validate:"required_if=Scheme kafka"`
}

type Subscriber struct {
	config SubscriberConfig
	topic  *cdkPubSub.Subscription
}

var _ ISubscriber = (*Subscriber)(nil)

func (c *Subscriber) Subscribe(ctx context.Context, handler SubscribeHandler) {
	// Loop on received messages. We can use a channel as a semaphore to limit how
	// many goroutines we have active at a time as well as wait on the goroutines
	// to finish before exiting.
	maxHandler := c.config.MaxHandler
	sem := make(chan struct{}, maxHandler)

receiverLoop:
	for {
		msg, err := c.topic.Receive(ctx)
		if err != nil {
			// Errors from Receive indicate that Receive will no longer succeed.
			log.Printf("Receiving message: %v", err)
			break
		}

		// Wait if there are too many active handle goroutines and acquire the
		// semaphore. If the context is canceled, stop waiting and start shutting
		// down.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break receiverLoop
		}

		// Handle the message in a new goroutine.
		go func() {
			defer func() {
				msg.Ack() // Messages must always be acknowledged with Ack.
				<-sem     // Release the semaphore.
			}()

			// Do work based on the message
			err := handler(ctx, &Message{
				LoggableID: msg.LoggableID,
				Body:       msg.Body,
			})

			if err != nil && errors.Is(err, ErrTriggerDoNackMessage) {
				fmt.Println("doing nack message")
				if msg.Nackable() {
					fmt.Println("nack message done")
					msg.Nack()
				}

				return
			}

		}()
	}

	// We're no longer receiving messages. Wait to finish handling any
	// unacknowledged messages by totally acquiring the semaphore.
	for n := 0; n < maxHandler; n++ {
		sem <- struct{}{}
	}
}

func (c *Subscriber) Shutdown(ctx context.Context) error {
	return c.topic.Shutdown(ctx)
}

var _ ISubscriber = (*Subscriber)(nil)

func NewSubscriber(conf SubscriberConfig) (*Subscriber, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	// The set of brokers in the Kafka cluster.
	const group = "group"
	// The Kafka client configuration to use.
	config := kafkapubsub.MinimalConfig()

	// Construct a *pubsub.Subscription, joining the consumer group "my-group"
	// and receiving messages from "my-topic".
	subscription, err := kafkapubsub.OpenSubscription(
		conf.Kafka.Brokers,
		config,
		group,
		[]string{conf.Kafka.Topic},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Subscriber{config: conf, topic: subscription}, nil
}
