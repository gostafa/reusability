package workerpool_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/gostafa/reusability/internal/shared/workerpool"
)

// Black-box: every task runs exactly once.
func TestRunAllTasks(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	err := workerpool.Run(context.Background(), 4, 100, func(int) error {
		count.Add(1)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if count.Load() != 100 {
		t.Fatalf("ran %d tasks, want 100", count.Load())
	}
}

// Black-box: the first task error surfaces.
func TestRunPropagatesError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("boom")

	err := workerpool.Run(context.Background(), 2, 10, func(i int) error {
		if i == 3 {
			return sentinel
		}

		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}
}

// Black-box: the worker count is bounded by the configured value and the task
// count.
func TestWorkersBounds(t *testing.T) {
	t.Parallel()

	if got := workerpool.Workers(2, 100); got != 2 {
		t.Errorf("Workers(2,100) = %d, want 2", got)
	}

	if got := workerpool.Workers(10, 3); got != 3 {
		t.Errorf("Workers(10,3) = %d, want 3 (task-bound)", got)
	}

	if got := workerpool.Workers(0, 3); got < 1 || got > 3 {
		t.Errorf("Workers(0,3) = %d, want within [1,3]", got)
	}
}
