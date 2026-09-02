package profiling_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/reusability/internal/infrastructure/profiling"
)

// Black-box: WriteHeap produces a non-empty heap profile.
func TestWriteHeap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "heap.prof")
	err := profiling.WriteHeap(path)
	if err != nil {
		t.Fatal(err)
	}

	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("heap profile not written: err=%v", err)
	}
}

// Black-box: WriteHeap surfaces file-creation errors.
func TestWriteHeapBadPath(t *testing.T) {
	err := profiling.WriteHeap(filepath.Join(t.TempDir(), "missing-dir", "heap.prof"))
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}
