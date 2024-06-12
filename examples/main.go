package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/initlevel5/workerpool"
	"go.uber.org/multierr"
)

const (
	numTasks         = 10
	numWorkers       = 2
	timeout          = 1 * time.Second
	hardWorkDuration = 2 * time.Second
)

func main() {
	var (
		err error
		wg  sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	defer func() {
		signal.Stop(signalCh)
		cancel()
	}()

	go func() {
		select {
		case <-signalCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	pool := workerpool.New(numWorkers, numTasks)

	errCh := make(chan error, numTasks)

	for i := 0; i < numTasks; i++ {
		id := i + 1

		if pool.AddTask(func() {
			defer wg.Done()

			ctxWithTimeout, ctxWithTimeoutCancel := context.WithTimeout(ctx, timeout)
			defer ctxWithTimeoutCancel()

			fmt.Printf("%d:started\n", id)

			select {
			case <-ctxWithTimeout.Done():
				errCh <- fmt.Errorf("%d: interrupted: %w", id, ctxWithTimeout.Err())
			case <-time.After(hardWorkDuration):
				fmt.Printf("%d:finished\n", id)
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
