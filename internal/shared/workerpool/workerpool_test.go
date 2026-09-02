package workerpool

import (
	"context"
	"errors"
	"runtime"
	"testing"
)

func TestRunIndexedResults(t *testing.T) {
	const n = 100

	results := make([]int, n)

	err := Run(context.Background(), 8, n, func(i int) error {
		results[i] = i * i

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for i, v := range results {
		if v != i*i {
			t.Fatalf("results[%d] = %d", i, v)
		}
	}
}

func TestRunFirstErrorByIndex(t *testing.T) {
	wantErr := errors.New("boom")

	err := Run(context.Background(), 4, 10, func(i int) error {
		if i == 3 || i == 7 {
			return wantErr
		}

		return nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v", err)
	}
}

func TestRunCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, 2, 1000, func(i int) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestRunZeroTasks(t *testing.T) {
	err := Run(context.Background(), 4, 0, func(int) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkers(t *testing.T) {
	maxProcs := runtime.GOMAXPROCS(0)
	if got := Workers(0, 1000); got != maxProcs {
		t.Fatalf("Workers(0, 1000) = %d, want %d", got, maxProcs)
	}

	if got := Workers(0, 1); got != 1 {
		t.Fatalf("Workers(0, 1) = %d, want 1", got)
	}

	if got := Workers(3, 1000); got != 3 {
		t.Fatalf("Workers(3, 1000) = %d, want 3", got)
	}

	if got := Workers(64, 2); got != 2 {
		t.Fatalf("Workers(64, 2) = %d, want 2", got)
	}

	if got := Workers(0, 0); got != 1 {
		t.Fatalf("Workers(0, 0) = %d, want 1", got)
	}
}
