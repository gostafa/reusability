package profiling

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// White-box: StartCPU writes a profile and the stop function closes it.
// (Not parallel — the CPU profiler is process-global.)
func TestStartCPUWritesProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.prof")

	stop, err := StartCPU(path)
	if err != nil {
		t.Fatal(err)
	}

	sum := 0
	for i := range 100000 {
		sum += i
	}

	_ = sum

	if err := stop(); err != nil {
		t.Fatal(err)
	}
	if err := stop(); err == nil {
		t.Fatal("expected a second stop to report that the profile is already closed")
	}

	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("cpu profile not written: err=%v", err)
	}
}

// White-box: StartCPU surfaces file-creation errors.
func TestStartCPUBadPath(t *testing.T) {
	_, err := StartCPU(filepath.Join(t.TempDir(), "missing-dir", "cpu.prof"))
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}

// StartCPU closes the newly-created file when the process-wide profiler is
// already active.
func TestStartCPUAlreadyActive(t *testing.T) {
	stop, err := StartCPU(filepath.Join(t.TempDir(), "first.prof"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := stop(); err != nil {
			t.Errorf("stop first profile: %v", err)
		}
	})

	second := filepath.Join(t.TempDir(), "second.prof")
	if _, err := StartCPU(second); err == nil {
		t.Fatal("expected an error when CPU profiling is already active")
	}

	file, err := os.OpenFile(second, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open second profile after failed start: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close second profile: %v", err)
	}
}

func TestCreateProfileFileWithoutDirectory(t *testing.T) {
	t.Chdir(t.TempDir())

	file, err := createProfileFile("cpu.prof")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWriteHeapWriteAndCloseErrors(t *testing.T) {
	origWrite, origClose := writeHeapProfile, closeFile
	t.Cleanup(func() { writeHeapProfile, closeFile = origWrite, origClose })

	sentinel := errors.New("heap write failed")
	writeHeapProfile = func(io.Writer) error { return sentinel }
	if err := WriteHeap(filepath.Join(t.TempDir(), "heap.prof")); !errors.Is(err, sentinel) {
		t.Fatalf("write error = %v", err)
	}

	writeHeapProfile = func(io.Writer) error { return nil }
	closeSentinel := errors.New("close failed")
	closeFile = func(*os.File) error { return closeSentinel }
	if err := WriteHeap(filepath.Join(t.TempDir(), "heap2.prof")); !errors.Is(err, closeSentinel) {
		t.Fatalf("close error = %v", err)
	}
}
