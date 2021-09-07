package msgservice

import (
	"context"
)

type Service interface {
	Process(ctx context.Context, task *Task) (out *TaskResult, err error)
}
