// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

import (
	"context"
	"fmt"
	"runtime"
)

// Run executes fn(i) for every i in [0, taskCount) on at most workers
// goroutines. It stops handing out new tasks once the context is canceled,
// and returns the context error then, or else the first task error by index.
func Run(ctx context.Context, cfg RunConfig) error {
	err := runOrEmpty(ctx, cfg)
	if err != nil {
		return fmt.Errorf("workerpool: %w", err)
	}

	return nil
}

func runOrEmpty(ctx context.Context, cfg RunConfig) error {
	err := pickPoolRunner(cfg)(ctx, cfg)
	if err != nil {
		return fmt.Errorf(errFmtOp, "pool", err)
	}

	return nil
}

func pickPoolRunner(cfg RunConfig) func(context.Context, RunConfig) error {
	if cfg.TaskCount == zero {
		return runEmptyPool
	}

	return runPool
}

func runEmptyPool(ctx context.Context, _ RunConfig) error {
	err := emptyPoolResult(ctx)
	if err != nil {
		return fmt.Errorf(errFmtOp, "empty pool", err)
	}

	return nil
}

func emptyPoolResult(ctx context.Context) error {
	err := contextError(ctx, "workerpool")
	if err != nil {
		return fmt.Errorf("workerpool empty: %w", err)
	}

	return nil
}

func runPool(ctx context.Context, cfg RunConfig) error {
	workerPool := newPool(cfg)
	workerPool.dispatch(ctx)
	workerPool.wait()

	err := workerPool.result(ctx)
	if err != nil {
		return fmt.Errorf("workerpool run: %w", err)
	}

	return nil
}

// Workers returns the effective worker count for taskCount tasks:
// min(GOMAXPROCS, taskCount) by default, min(configured, taskCount) when a
// positive override is given.
func Workers(configured, taskCount int) int {
	workers := min(runtime.GOMAXPROCS(zero), taskCount)

	if configured > zero {
		workers = min(configured, taskCount)
	}

	return max(workers, one)
}

func contextError(ctx context.Context, prefix string) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf(errFmtOp, prefix, err)
	}

	return nil
}

func newPool(cfg RunConfig) *pool {
	workers := min(max(cfg.Workers, one), cfg.TaskCount)
	workerPool := &pool{
		tasks: make(chan int),
		errs:  make([]error, cfg.TaskCount),
	}

	workerPool.wg.Add(workers)

	for range workers {
		go workerPool.worker(cfg.Fn)
	}

	return workerPool
}

func (workerPool *pool) dispatch(ctx context.Context) {
	for i := zero; i < len(workerPool.errs); i++ {
		select {
		case workerPool.tasks <- i:
		case <-ctx.Done():
			return
		}
	}
}

func (workerPool *pool) result(ctx context.Context) error {
	err := contextError(ctx, "workerpool result")
	if err != nil {
		return fmt.Errorf("pool result: %w", err)
	}

	for i := range workerPool.errs {
		if workerPool.errs[i] != nil {
			return fmt.Errorf("workerpool task %d: %w", i, workerPool.errs[i])
		}
	}

	return nil
}

func (workerPool *pool) wait() {
	close(workerPool.tasks)
	workerPool.wg.Wait()
}

func (workerPool *pool) worker(fn func(int) error) {
	defer workerPool.wg.Done()

	for i := range workerPool.tasks {
		workerPool.errs[i] = fn(i)
	}
}
