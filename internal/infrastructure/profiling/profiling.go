package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

// Seams so tests can force write/close failures.
var (
	writeHeapProfile = pprof.WriteHeapProfile
	closeFile        = func(f *os.File) error { return f.Close() }
)

// StartCPU begins CPU profiling into path and returns a stop function that
// finishes the profile and closes the file.
func StartCPU(path string) (stop func() error, err error) {
	f, err := createProfileFile(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		_ = closeFile(f)

		return nil, fmt.Errorf("start cpu profile: %w", err)
	}

	return func() error {
		pprof.StopCPUProfile()

		err := closeFile(f)
		if err != nil {
			return fmt.Errorf("close cpu profile: %w", err)
		}

		return nil
	}, nil
}

// WriteHeap writes a heap profile to path after forcing a garbage collection
// so the profile reflects live allocations.
func WriteHeap(path string) error {
	f, err := createProfileFile(path)
	if err != nil {
		return fmt.Errorf("create memory profile: %w", err)
	}

	runtime.GC()

	if err := writeHeapProfile(f); err != nil {
		_ = closeFile(f)

		return fmt.Errorf("write memory profile: %w", err)
	}

	if err := closeFile(f); err != nil {
		return fmt.Errorf("close memory profile: %w", err)
	}

	return nil
}

func createProfileFile(path string) (*os.File, error) {
	dir, name := filepath.Split(path)
	if dir == "" {
		dir = "."
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open profile directory: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	return root.Create(name)
}
