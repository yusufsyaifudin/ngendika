package msgservice

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/pkg/pubsub"
)

type DispatcherConfig struct {
	Publisher queue.IPublisher `validate:"required"`
}

type Dispatcher struct {
	Config DispatcherConfig
}

var _ Service = (*Dispatcher)(nil)

func NewDispatcher(conf DispatcherConfig) (*Dispatcher, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &Dispatcher{Config: conf}, nil
}

func (d *Dispatcher) Process(ctx context.Context, task *Task) (out *TaskResult, err error) {
	err = validator.New().Struct(task)
	if err != nil {
		return
	}

	data, err := json.Marshal(task)
	if err != nil {
		return
	}

	err = d.Config.Publisher.Publish(ctx, &queue.Message{
		Body: data,
	})

	if err != nil {
		return
	}

	// build result
	out = &TaskResult{
		TaskID:      task.TaskID,
		AppClientID: task.AppClientID,
	}
	return
}
