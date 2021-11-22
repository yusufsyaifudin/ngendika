package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/multierr"
)

var (
	ErrPreExecute = errors.New("pre-execute job error")
	ErrExecute    = errors.New("execute job error")
)

// Job holds all information regarding the Job
type Job interface {
	// ID return uint64 unique identifier of the job
	ID() uint64

	// Context to tracks down all Job information that important.
	Context() context.Context

	// PreExecute called before Execute, when error Execute never be called.
	// PostExecute always called after PreExecute or Execute is done.
	PreExecute() error

	// Execute is the real logic of the Job.
	Execute() error

	// PostExecute called after Execute is done.
	// When Execute return error, it will pass to PostExecute, otherwise it returns nil.
	PostExecute(err error)
}

type Service interface {
	AddJob(job Job)
	WaitJob(job Job)
}

type Worker struct {
	waitGroup   *sync.WaitGroup
	JobQueue    chan Job
	JobQueueNum int64
	Logger      Logger
	Stop        chan bool
}

var _ Service = (*Worker)(nil)

func NewWorker(num, maxJob int) *Worker {
	if num < 1 {
		num = 1
	}

	if maxJob < 1 {
		maxJob = 1
	}

	wg := &sync.WaitGroup{}

	jobs := make(chan Job, maxJob)

	w := &Worker{
		waitGroup: wg,
		JobQueue:  jobs,
		Logger:    new(stdOut),
		Stop:      make(chan bool),
	}

	for i := 0; i < num; i++ {
		go w.worker(i + 1)
	}

	return w
}

func (w *Worker) worker(id int) {
	go func() {
		for {
			select {
			case job := <-w.JobQueue:
				// fmt.Println("worker", id, "started  job", job.ID())
				if job == nil {
					continue
				}

				t0 := time.Now()
				var _err error
				_err = job.PreExecute()
				if _err != nil {
					_err = multierr.Append(_err, ErrPreExecute)
					job.PostExecute(_err)
					atomic.AddInt64(&w.JobQueueNum, -1)
					w.waitGroup.Done()
					continue
				}

				_err = job.Execute()
				if _err != nil {
					_err = multierr.Append(_err, ErrExecute)
				}

				job.PostExecute(_err)
				atomic.AddInt64(&w.JobQueueNum, -1)
				w.waitGroup.Done()

				w.Logger.Info(
					job.Context(),
					fmt.Sprintf("worker %d, job id %d, ongoing queue %d, duration %s",
						id, job.ID(), atomic.LoadInt64(&w.JobQueueNum), time.Since(t0).String(),
					),
				)

			case <-w.Stop:
				close(w.JobQueue)
				return
			}
		}
	}()
}

func (w *Worker) AddJob(job Job) {
	if job == nil {
		return
	}

	w.waitGroup.Add(1)
	w.JobQueue <- job
	atomic.AddInt64(&w.JobQueueNum, 1)
}

func (w *Worker) WaitJob(job Job) {
	doneChan := make(chan struct{})
	job.Execute()
	close(doneChan)
	<-doneChan
}

// Done ensures all registered Job is done before stop the worker.
func (w *Worker) Done() {
	for atomic.LoadInt64(&w.JobQueueNum) > 0 {
		continue
	}

	w.waitGroup.Wait()
	w.Stop <- true
}
