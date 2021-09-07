package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yusufsyaifudin/ngendika/pkg/worker"
)

type Job struct {
	id         uint64
	preExecErr error
	execErr    error
}

func (s *Job) ID() uint64 {
	return s.id
}

func (s *Job) Context() context.Context {
	return context.Background()
}

func (s *Job) PreExecute() error {
	return s.preExecErr
}

func (s *Job) Execute() error {
	time.Sleep(100 * time.Millisecond)
	return s.execErr
}

func (s *Job) PostExecute(err error) {
	return
}

func TestNewWorker(t *testing.T) {
	t.Run("worker lower than 1", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(0, 100)
		defer dispatcher.Done()
		dispatcher.AddJob(&Job{id: 1})
	})

	t.Run("max job lower than 1", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(4, 0)
		defer dispatcher.Done()
		dispatcher.AddJob(&Job{id: 1})
	})

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(4, 100)
		defer dispatcher.Done()

		for i := 0; i < 100; i++ {
			id := uint64(i)
			dispatcher.AddJob(&Job{id: id})
		}
	})

	t.Run("job is nil", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(4, 100)
		defer dispatcher.Done()

		for i := 0; i < 100; i++ {
			dispatcher.AddJob(nil)
		}
	})

	t.Run("pre execute is error", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(4, 100)
		defer dispatcher.Done()

		for i := 0; i < 100; i++ {
			id := uint64(i)
			dispatcher.AddJob(&Job{id: id, preExecErr: errors.New("shit happen")})
		}
	})

	t.Run("execute is error", func(t *testing.T) {
		t.Parallel()

		dispatcher := worker.NewWorker(4, 100)
		defer dispatcher.Done()

		for i := 0; i < 100; i++ {
			id := uint64(i)
			dispatcher.AddJob(&Job{id: id, execErr: errors.New("shit happen")})
		}
	})

	// // Handle sigterm and await termChan signal
	// termChan := make(chan os.Signal)
	// signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	//
	// <-termChan // Blocks here until interrupted
}

func BenchmarkNewWorker(b *testing.B) {
	dispatcher := worker.NewWorker(8, 100)
	defer dispatcher.Done()

	for i := 0; i < b.N; i++ {
		id := uint64(i)
		dispatcher.AddJob(&Job{id: id})
	}
}
