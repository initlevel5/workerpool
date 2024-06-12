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

var ErrClosed = errors.New("closed")

type workerPool struct {
	tasks  chan func()
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
		tasks: make(chan func(), queueLen),
	}

	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)

		go worker(&pool.wg, pool.tasks)
	}

	return pool
}

func worker(wg *sync.WaitGroup, tasks <-chan func()) {
	defer wg.Done()

	for {
		select {
		//case <-ctx.Done():
		//	return
		case task, ok := <-tasks:
			if task == nil || !ok {
				return
			}
			task()
		}
	}
}

func (p *workerPool) AddTask(task func()) bool {
	return p.addTask(task, false)
}

func (p *workerPool) MustAddTask(task func()) bool {
	return p.addTask(task, true)
}

func (p *workerPool) addTask(task func(), must bool) bool {
	if task == nil {
		return false
	}

	if p.IsClosed() {
		if must {
			panic(ErrClosed)
		}
		return false
	}

	if !must {
		select {
		case p.tasks <- task:
			return true
		default:
			return false
		}
	}

	p.tasks <- task

	return true
}

func (p *workerPool) IsClosed() (closed bool) {
	p.mu.RLock()
	closed = p.closed
	p.mu.RUnlock()

	return
}

func (p *workerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.tasks)
}

func (p *workerPool) Wait() {
	p.wg.Wait()
}
