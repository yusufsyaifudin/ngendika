package msgservice

import (
	"context"
	"fmt"

	"github.com/yusufsyaifudin/ngendika/pkg/worker"
)

type Job struct {
	ctx context.Context
	id  uint64
	f   func() (interface{}, error)
}

func (j *Job) ID() uint64 {
	return j.id
}

func (j *Job) Context() context.Context {
	return j.ctx
}

func (j *Job) PreExecute() error {
	return nil
}

func (j *Job) Execute() error {
	out, err := j.f()
	fmt.Println(out)
	if err != nil {
		return err
	}

	return err
}

func (j *Job) PostExecute(err error) {
	fmt.Println("job", j.id, "err", err)
}

func newJob(ctx context.Context, id uint64, f func() (interface{}, error)) *Job {
	return &Job{
		ctx: ctx,
		id:  id,
		f:   f,
	}
}

var _ worker.Job = (*Job)(nil)
