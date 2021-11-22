package msgservice

import (
	"context"
)

type Process func(ctx context.Context, task *Task) (out *TaskResult, err error)

type Service interface {
	Process() Process
}
