package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/initlevel5/workerpool"
)

func main() {
	const (
		nTasks              = 10
		nWorkers            = 2
		timeoutSec          = 20
		hardWorkDurationSec = 2
	)

	var wg sync.WaitGroup

	ctx := context.Background()

	pool := workerpool.New(nWorkers, nTasks)

	pool.Run(ctx)

	taskFunc := func(ctx context.Context, id int, data interface{}) error {
		fmt.Printf("%d:start work\n", id)

		select {
		case <-ctx.Done():
			return fmt.Errorf("interrupted: %w", ctx.Err())
		case <-time.After(hardWorkDurationSec * time.Second):
			return nil
		}
	}

	withTimeout, cancel := context.WithTimeout(ctx, timeoutSec*time.Second)
	defer cancel()

	for i := 0; i < nTasks; i++ {
		task := workerpool.NewTask(withTimeout, i+1, i, taskFunc)

		_ = pool.AddTask(task)

		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case <-task.Done():
				if err := task.Err(); err != nil {
					fmt.Printf("%d:error: %v\n", task.Id(), err)
					return
				}
				fmt.Printf("%d:done\n", task.Id())
			case <-ctx.Done():
				fmt.Println(ctx.Err())
			}
		}()
	}

	wg.Wait()

	pool.Stop()
	pool.Wait()
}
