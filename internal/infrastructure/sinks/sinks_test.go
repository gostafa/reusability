// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

// Black-box: FileSink round-trips a report body through the port.
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
