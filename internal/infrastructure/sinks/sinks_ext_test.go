package sinks_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/internal/infrastructure/sinks"
)

// Both adapters satisfy the reporting outbound port.
var (
	_ outbound.Sink = sinks.StdoutSink{}
	_ outbound.Sink = sinks.FileSink{}
)

// Black-box: FileSink round-trips a report body through the port.
func TestFileSinkRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "report.json")

	var s outbound.Sink = sinks.FileSink{Path: path}

	w, err := s.Open()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(w, `{"ok":true}`); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != `{"ok":true}` {
		t.Fatalf("round-trip = %q", data)
	}
}

// Black-box: StdoutSink writes to standard output (redirected here to a pipe).
func TestStdoutSinkWritesToStdout(t *testing.T) {
	orig := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w
	defer func() { os.Stdout = orig }()

	out, err := sinks.StdoutSink{}.Open()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(out, "hello stdout"); err != nil {
		t.Fatal(err)
	}

	if err := out.Close(); err != nil { // flushes to the pipe
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	data, _ := io.ReadAll(r)
	if string(data) != "hello stdout" {
		t.Fatalf("stdout captured = %q", data)
	}
}
