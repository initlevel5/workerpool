package workerpool

import (
	"context"
	"sync"
)

const (
	numWorkersMax     = 16
	numWorkersDefault = 2
	queueLenMax       = 1024
	queueLenDefault   = 16
)

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

func (t Task) Context() context.Context {
	return t.ctx
}

func (t Task) Id() int {
	return t.id
}

func (t Task) Data() interface{} {
	return t.data
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
	stop, done chan struct{}
}

func NewWorkerPool(numWorkers, queueLen int) *workerPool {
	if numWorkers <= 0 || numWorkers > numWorkersMax {
		numWorkers = numWorkersDefault
	}

	if queueLen <= 0 || queueLen > queueLenMax {
		queueLen = queueLenDefault
	}

	return &workerPool{
		numWorkers: numWorkers,
		tasks:      make(chan *Task, queueLen),
		stop:       make(chan struct{}, 1),
		done:       make(chan struct{}, 1),
	}
}

func (wp *workerPool) Run(ctx context.Context) {
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
					case <-wp.stop:
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
}

func (wp workerPool) Stop() {
	close(wp.stop)
}

func (wp workerPool) Wait() {
	<-wp.done
}

func (wp *workerPool) AddTask(t *Task) {
	wp.tasks <- t
}
