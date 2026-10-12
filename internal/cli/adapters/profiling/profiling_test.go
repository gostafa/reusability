// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Black-box: WriteHeap produces a non-empty heap profile.
func TestCPUProfileLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.prof")
	stop, err := StartCPU(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if stop != nil {
			_ = stop()
		}
	})
	if _, err := StartCPU(filepath.Join(t.TempDir(), "second.prof")); err == nil {
		t.Fatal("overlapping profiles accepted")
	}
	sentinel := errors.New("close failed")
	deps := defaultProfileDeps()
	deps.closeFile = func(file *os.File) error { _ = file.Close(); return sentinel }
	if _, err := startCPUWith(filepath.Join(t.TempDir(), "third.prof"), deps); err == nil {
		t.Fatal("expected start and close failure")
	}
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	stop = nil
	if _, err := StartCPU(filepath.Join(t.TempDir(), "missing", "cpu.prof")); err == nil {
		t.Fatal("invalid CPU path accepted")
	}
	stop, err = startCPUWith(filepath.Join(t.TempDir(), "closing.prof"), deps)
	if err != nil {
		t.Fatal(err)
	}
	if err := stop(); !errors.Is(err, sentinel) {
		t.Fatalf("stop = %v", err)
	}
	stop = nil
}

func TestProfileIOFailures(t *testing.T) {
	sentinel := errors.New("profile failure")
	closeRootWith(func() error { return sentinel })
	for _, failClose := range []bool{false, true} {
		deps := profileDeps{
			writeHeap: func(*os.File) error { return sentinel },
			closeFile: func(file *os.File) error {
				_ = file.Close()
				if failClose {
					return sentinel
				}
				return nil
			},
		}
		if err := writeHeapWith(
			filepath.Join(t.TempDir(), "heap.prof"),
			deps,
		); !errors.Is(
			err,
			sentinel,
		) {
			t.Fatalf("write = %v", err)
		}
		deps.writeHeap = func(*os.File) error { return nil }
		if err := writeHeapWith(
			filepath.Join(t.TempDir(), "heap.prof"),
			deps,
		); (err != nil) != failClose {
			t.Fatalf("close = %v", err)
		}
	}
	if _, err := createProfileFile(t.TempDir()); err == nil {
		t.Fatal("directory accepted as file")
	}
	file, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := closeFile(file); err == nil {
		t.Fatal("closed file accepted")
	}
	if err := writeHeapProfile(file); err == nil {
		t.Fatal("write to closed file accepted")
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := root.Close(); err != nil {
		t.Fatal(err)
	}
	closeRoot(root)
	if dir, name := splitProfilePath("cpu.prof"); dir != "." || name != "cpu.prof" {
		t.Fatalf("split = %q/%q", dir, name)
	}
}

func TestWriteHeap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "heap.prof")
	err := WriteHeap(path)
	if err != nil {
		t.Fatal(err)
	}

	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("heap profile not written: err=%v", err)
	}
}

// Black-box: WriteHeap surfaces file-creation errors.
func TestWriteHeapBadPath(t *testing.T) {
	err := WriteHeap(filepath.Join(t.TempDir(), "missing-dir", "heap.prof"))
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}
