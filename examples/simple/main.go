package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/initlevel5/workerpool"
)

func main() {
	var (
		wg      sync.WaitGroup
		counter atomic.Int32
	)

	numWorkers := runtime.NumCPU()
	numTasks := numWorkers * 2

	pool := workerpool.New(numWorkers, numTasks)
	defer pool.Close()

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		pool.MustAddTask(func() {
			counter.Add(1)
			wg.Done()
		})
	}

	wg.Wait()

	fmt.Println(counter.Load())
}
