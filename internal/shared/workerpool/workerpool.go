package workerpool

import (
	"context"
	"runtime"
	"sync"
)

// Workers returns the effective worker count for taskCount tasks:
// min(GOMAXPROCS, taskCount) by default, min(configured, taskCount) when a
// positive override is given.
func Workers(configured, taskCount int) int {
	workers := min(runtime.GOMAXPROCS(0), taskCount)
	if configured > 0 {
		workers = min(configured, taskCount)
	}

	return max(workers, 1)
}

// Run executes fn(i) for every i in [0, taskCount) on at most workers
// goroutines. It stops handing out new tasks once the context is cancelled,
// and returns the context error then, or else the first task error by index.
func Run(ctx context.Context, workers, taskCount int, fn func(i int) error) error {
	if taskCount == 0 {
		return ctx.Err()
	}

	workers = min(max(workers, 1), taskCount)

	tasks := make(chan int)
	errs := make([]error, taskCount)

	var wg sync.WaitGroup

	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			for i := range tasks {
				errs[i] = fn(i)
			}
		}()
	}

	var stopped bool
	for i := 0; i < taskCount && !stopped; i++ {
		select {
		case tasks <- i:
		case <-ctx.Done():
			stopped = true
		}
	}

	close(tasks)
	wg.Wait()

	err := ctx.Err()
	if err != nil {
		return err
	}

	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
