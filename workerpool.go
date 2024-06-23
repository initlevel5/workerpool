package workerpool

import (
	"errors"
	"sync"
)

const (
	numWorkersMax     = 16
	numWorkersDefault = 2
	queueLenMax       = 1024
	queueLenDefault   = 16
)

var (
	ErrClosed          = errors.New("closed")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrFull            = errors.New("task queue is full")
)

type Task func()

type workerPool struct {
	taskCh chan Task
	wg     sync.WaitGroup
	closed bool
	mu     sync.RWMutex
}

func New(numWorkers, queueLen int) *workerPool {
	if numWorkers <= 0 || numWorkers > numWorkersMax {
		numWorkers = numWorkersDefault
	}

	if queueLen <= 0 || queueLen > queueLenMax {
		queueLen = queueLenDefault
	}

	pool := &workerPool{
		taskCh: make(chan Task, queueLen),
	}

	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)

		go worker(&pool.wg, pool.taskCh)
	}

	return pool
}

func worker(wg *sync.WaitGroup, taskCh <-chan Task) {
	defer wg.Done()

	for {
		select {
		//case <-ctx.Done():
		//	return
		case task, ok := <-taskCh:
			if task == nil || !ok {
				return
			}
			task()
		}
	}
}

func (p *workerPool) AddTask(task Task) error {
	return p.addTask(task)
}

func (p *workerPool) MustAddTask(task Task) {
	if err := p.addTask(task); err != nil {
		panic(err)
	}
}

func (p *workerPool) addTask(task Task) error {
	if task == nil {
		return ErrInvalidArgument
	}

	if p.IsClosed() {
		return ErrClosed
	}

	select {
	case p.taskCh <- task:
		return nil
	default:
		return ErrFull
	}
}

func (p *workerPool) IsClosed() (closed bool) {
	p.mu.RLock()
	closed = p.closed
	p.mu.RUnlock()

	return
}

func (p *workerPool) Close() {
	p.close()
	p.wg.Wait()
}

func (p *workerPool) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.taskCh)
}
