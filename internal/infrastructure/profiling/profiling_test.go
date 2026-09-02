// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

import (
	"os"
	"path/filepath"
	"testing"
)

// Black-box: WriteHeap produces a non-empty heap profile.
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
