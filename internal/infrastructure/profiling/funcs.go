// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
)

type (
	profileDeps struct {
		writeHeap func(*os.File) error
		closeFile func(*os.File) error
	}
)

const (
	errStartCPUProfile  = "start cpu profile: %w"
	errWriteHeapProfile = "write heap profile: %w"
)

func defaultProfileDeps() profileDeps {
	return profileDeps{
		writeHeap: writeHeapProfile,
		closeFile: closeFile,
	}
}

func closeFile(file *os.File) error {
	err := file.Close()
	if err != nil {
		return fmt.Errorf("close profile file: %w", err)
	}

	return nil
}

func writeHeapProfile(file *os.File) error {
	err := pprof.WriteHeapProfile(file)
	if err != nil {
		return fmt.Errorf(errWriteHeapProfile, err)
	}

	return nil
}

// StartCPU begins CPU profiling into path and returns a stop function that
// finishes the profile and closes the file.
func StartCPU(path string) (stop func() error, err error) {
	stop, err = startCPUWith(path, defaultProfileDeps())
	if err != nil {
		return nil, fmt.Errorf(errStartCPUProfile, err)
	}

	return stop, nil
}

func startCPUWith(path string, deps profileDeps) (stop func() error, err error) {
	file, err := createProfileFile(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}

	startErr := beginCPUProfile(file)
	if startErr != nil {
		failErr := failStartCPU(file, deps, startErr)

		return nil, fmt.Errorf(errStartCPUProfile, failErr)
	}

	return endCPUProfile(file, deps), nil
}

func beginCPUProfile(file *os.File) error {
	err := pprof.StartCPUProfile(file)
	if err != nil {
		return fmt.Errorf("begin cpu profile: %w", err)
	}

	return nil
}

func failStartCPU(file *os.File, deps profileDeps, startErr error) error {
	closeErr := deps.closeFile(file)
	if closeErr != nil {
		return fmt.Errorf("start cpu profile: %w (close: %v)", startErr, closeErr)
	}

	return fmt.Errorf(errStartCPUProfile, startErr)
}

func endCPUProfile(file *os.File, deps profileDeps) func() error {
	return func() error {
		pprof.StopCPUProfile()

		closeErr := deps.closeFile(file)
		if closeErr != nil {
			return fmt.Errorf("close cpu profile: %w", closeErr)
		}

		return nil
	}
}

// WriteHeap writes a heap profile to path. The profile may include unreclaimed
// garbage because no explicit GC is run before capture.
func WriteHeap(path string) error {
	err := writeHeapWith(path, defaultProfileDeps())
	if err != nil {
		return fmt.Errorf(errWriteHeapProfile, err)
	}

	return nil
}

func writeHeapWith(path string, deps profileDeps) error {
	file, err := createProfileFile(path)
	if err != nil {
		return fmt.Errorf("create memory profile: %w", err)
	}

	writeErr := writeHeapBody(file, deps)
	if writeErr != nil {
		return fmt.Errorf("write heap body: %w", writeErr)
	}

	closeErr := deps.closeFile(file)
	if closeErr != nil {
		return fmt.Errorf("close memory profile: %w", closeErr)
	}

	return nil
}

func writeHeapBody(file *os.File, deps profileDeps) error {
	writeErr := deps.writeHeap(file)
	if writeErr != nil {
		failErr := failWriteHeap(file, deps, writeErr)

		return fmt.Errorf(errWriteHeapProfile, failErr)
	}

	return nil
}

func failWriteHeap(file *os.File, deps profileDeps, writeErr error) error {
	closeErr := deps.closeFile(file)
	if closeErr != nil {
		return fmt.Errorf("write memory profile: %w (close: %v)", writeErr, closeErr)
	}

	return fmt.Errorf("write memory profile: %w", writeErr)
}

func createProfileFile(path string) (*os.File, error) {
	dir, name := splitProfilePath(path)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open profile directory: %w", err)
	}

	defer closeRoot(root)

	file, createErr := createInRoot(root, name)
	if createErr != nil {
		return nil, fmt.Errorf("create profile in root: %w", createErr)
	}

	return file, nil
}

func splitProfilePath(path string) (dir, name string) {
	dir, name = filepath.Split(path)

	if dir == "" {
		dir = "."
	}

	return dir, name
}

func createInRoot(root *os.Root, name string) (*os.File, error) {
	file, createErr := root.Create(name)
	if createErr != nil {
		return nil, fmt.Errorf("create profile file: %w", createErr)
	}

	return file, nil
}

func closeRoot(root *os.Root) {
	closeErr := root.Close()
	if closeErr != nil {
		return
	}
}
