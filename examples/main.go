package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/initlevel5/workerpool"
	"go.uber.org/multierr"
)

const (
	numTasks            = 10
	numWorkers          = 2
	timeoutSec          = 10
	hardWorkDurationSec = 2
)

func main() {
	var (
		err error
		wg  sync.WaitGroup
	)

	ctx := context.Background()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()

	pool := workerpool.New(ctx, numWorkers, numTasks)

	errCh := make(chan error, numTasks)

	for i := 0; i < numTasks; i++ {
		id := i + 1

		if pool.AddTask(func() {
			defer wg.Done()

			fmt.Printf("%d:start work\n", id)

			select {
			case <-ctxWithTimeout.Done():
				errCh <- fmt.Errorf("%d: interrupted: %w", id, ctxWithTimeout.Err())
			case <-time.After(hardWorkDurationSec * time.Second):
				return
			}
		}) {
			wg.Add(1)
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for e := range errCh {
		err = multierr.Append(err, e)
	}

	if err != nil {
		fmt.Println(err)
	}

	pool.Stop()
	pool.Wait()
}
