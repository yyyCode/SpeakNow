package queue

import (
	"context"
	"errors"
	"sync"

	"speaknow/internal/domain"
)

var ErrQueueFull = errors.New("task queue is full")

type Task struct {
	ID     string
	Audio  []byte
	Opts   domain.RecognizeOpts
	UserID string
	Result chan *TaskResult
}

type TaskResult struct {
	Response *domain.Result
	Err      error
}

type TaskQueue struct {
	tasks   chan *Task
	workers int
	wg      sync.WaitGroup
	recognize func(ctx context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error)
}

func NewTaskQueue(workers, buffer int, recognize func(context.Context, []byte, domain.RecognizeOpts) (*domain.Result, error)) *TaskQueue {
	q := &TaskQueue{
		tasks:     make(chan *Task, buffer),
		workers:   workers,
		recognize: recognize,
	}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
	return q
}

func (q *TaskQueue) Submit(task *Task) error {
	select {
	case q.tasks <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *TaskQueue) worker() {
	defer q.wg.Done()
	for task := range q.tasks {
		result, err := q.recognize(context.Background(), task.Audio, task.Opts)
		task.Result <- &TaskResult{Response: result, Err: err}
		close(task.Result)
	}
}

func (q *TaskQueue) Close() {
	close(q.tasks)
	q.wg.Wait()
}

func (q *TaskQueue) Depth() int {
	return len(q.tasks)
}
