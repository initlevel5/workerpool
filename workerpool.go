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

func (p *workerPool) AddTask(task Task) bool {
	return p.addTask(task, false)
}

func (p *workerPool) MustAddTask(task Task) bool {
	return p.addTask(task, true)
}

func (p *workerPool) addTask(task Task, must bool) bool {
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
		case p.taskCh <- task:
			return true
		default:
			return false
		}
	}

	p.taskCh <- task

	return true
}

func (p *workerPool) IsClosed() (closed bool) {
	p.mu.RLock()
	closed = p.closed
	p.mu.RUnlock()

	return
}

func (p *workerPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.taskCh)
}

func (p *workerPool) Wait() {
	p.wg.Wait()
}
