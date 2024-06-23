package workerpool

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	const (
		numWorkers = 2
		numTasks   = numWorkers * 10
	)

	var (
		wg             sync.WaitGroup
		tasksDoneCount atomic.Int32
	)

	pool := New(numWorkers, numTasks)
	defer pool.Close()

	for i := 0; i < numTasks; i++ {
		wg.Add(1)

		pool.MustAddTask(func() {
			tasksDoneCount.Add(1)
			wg.Done()
		})
	}

	wg.Wait()

	actual := tasksDoneCount.Load()

	if actual != numTasks {
		t.Errorf("actual: %d, expected %d", actual, numTasks)
	}
}
