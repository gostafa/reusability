package sinks

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

// White-box: stdoutStream.Close flushes buffered output (without closing the
// underlying descriptor, which the process owns).
func TestStdoutStreamCloseFlushes(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "stream")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = f.Close() }()

	s := stdoutStream{bufio.NewWriter(f)}
	if _, err := s.Write([]byte("buffered")); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f.Name())
	if string(data) != "buffered" {
		t.Fatalf("Close did not flush: %q", data)
	}
}

// White-box: FileSink.Open creates and truncates the destination file.
func TestFileSinkOpenTruncates(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "out.txt")
	if err := os.WriteFile(path, []byte("stale, longer content"), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := FileSink{Path: path}.Open()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := w.Write([]byte("new")); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Fatalf("file not truncated: %q", data)
	}
}
