package workerpool

import (
	"context"
	"errors"
	"sync"
)

const (
	numWorkersMax     = 16
	numWorkersDefault = 2
	queueLenMax       = 1024
	queueLenDefault   = 16
)

var ErrClosed = errors.New("closed")

type TaskFunc func(context.Context, int, interface{}) error

type Task struct {
	ctx  context.Context
	id   int
	data interface{}
	f    TaskFunc
	done chan struct{}
	err  error
}

func NewTask(ctx context.Context, id int, data interface{}, f TaskFunc) *Task {
	return &Task{
		ctx:  ctx,
		id:   id,
		data: data,
		f:    f,
		done: make(chan struct{}, 1),
	}
}

func (t Task) Id() int {
	return t.id
}

func (t Task) Done() <-chan struct{} {
	return t.done
}

func (t Task) Err() error {
	return t.err
}

type workerPool struct {
	numWorkers int
	tasks      chan *Task
	done       chan struct{}
	once       sync.Once
	closed     bool
	mu         sync.Mutex
}

func New(numWorkers, queueLen int) *workerPool {
	if numWorkers <= 0 || numWorkers > numWorkersMax {
		numWorkers = numWorkersDefault
	}

	if queueLen <= 0 || queueLen > queueLenMax {
		queueLen = queueLenDefault
	}

	return &workerPool{
		numWorkers: numWorkers,
		tasks:      make(chan *Task, queueLen),
		done:       make(chan struct{}, 1),
	}
}

func (wp *workerPool) Run(ctx context.Context) {
	wp.once.Do(func() {
		go func() {
			defer close(wp.done)

			var wg sync.WaitGroup

			for i := 0; i < wp.numWorkers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for {
						select {
						case <-ctx.Done():
							return
						case task, ok := <-wp.tasks:
							if !ok {
								return
							}
							task.err = task.f(task.ctx, task.id, task.data)
							close(task.done)
						}
					}
				}()
			}

			wg.Wait()
		}()
	})
}

func (wp *workerPool) IsClosed() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	return wp.closed
}

func (wp *workerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.closed {
		close(wp.tasks)
		wp.closed = true
	}
}

func (wp *workerPool) Wait() {
	<-wp.done
}

func (wp *workerPool) AddTask(t *Task) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.closed {
		return ErrClosed
	}

	wp.tasks <- t

	return nil
}
