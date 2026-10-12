// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

// Black-box: FileSink round-trips a report body through the port.
type failedWriter struct{ err error }

func (writer failedWriter) Write([]byte) (int, error) { return 0, writer.err }

func TestSinkFailures(t *testing.T) {
	if _, err := OpenFile(FileSink{Path: t.TempDir()}); err == nil {
		t.Fatal("directory accepted")
	}
	old := os.Stdout
	os.Stdout = nil
	t.Cleanup(func() { os.Stdout = old })
	if _, err := OpenStdout(); err == nil {
		t.Fatal("nil stdout accepted")
	}
	sentinel := errors.New("write failure")
	stream := stdoutStream{writer: bufio.NewWriterSize(failedWriter{sentinel}, 1)}
	if _, err := stream.Write(bytes.Repeat([]byte("x"), 1024)); !errors.Is(err, sentinel) {
		t.Fatalf("write = %v", err)
	}
	stream = stdoutStream{writer: bufio.NewWriter(failedWriter{sentinel})}
	if _, err := stream.Write([]byte("small")); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); !errors.Is(err, sentinel) {
		t.Fatalf("flush = %v", err)
	}
}

func TestFileSinkRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "report.json")

	sink := outbound.NewSink(func() (*outbound.Stream, error) {
		return OpenFile(FileSink{Path: path})
	})

	stream, err := outbound.Open(sink)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(stream, `{"ok":true}`); err != nil {
		t.Fatal(err)
	}

	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != `{"ok":true}` {
		t.Fatalf("round-trip = %q", data)
	}
}

// Black-box: StdoutSink writes to standard output (redirected here to a pipe).
func TestStdoutSinkWritesToStdout(t *testing.T) {
	orig := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer
	defer func() { os.Stdout = orig }()

	stream, err := OpenStdout()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(stream, "hello stdout"); err != nil {
		t.Fatal(err)
	}

	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello stdout" {
		t.Fatalf("stdout captured = %q", data)
	}
}
