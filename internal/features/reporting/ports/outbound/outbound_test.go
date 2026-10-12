// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type nopWriteCloser struct{ io.Writer }

type failedWriteCloser struct{ err error }

func (writer failedWriteCloser) Write([]byte) (int, error) { return 0, writer.err }
func (writer failedWriteCloser) Close() error              { return writer.err }

func TestStreamPropagatesFailures(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("sink failure")
	if _, err := Open(
		NewSink(func() (*Stream, error) { return nil, sentinel }),
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("open = %v", err)
	}
	stream := NewStream(failedWriteCloser{sentinel})
	if _, err := stream.Write([]byte("report")); !errors.Is(err, sentinel) {
		t.Fatalf("write = %v", err)
	}
	if err := stream.Close(); !errors.Is(err, sentinel) {
		t.Fatalf("close = %v", err)
	}
}

func (nopWriteCloser) Close() error { return nil }

// White-box: NewSink yields a usable stream.
func TestSinkContract(t *testing.T) {
	t.Parallel()

	sink := NewSink(func() (*Stream, error) {
		return NewStream(nopWriteCloser{io.Discard}), nil
	})

	stream, err := Open(sink)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(stream, "hello"); err != nil {
		t.Fatal(err)
	}

	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
}

// Black-box: an external open function can capture what the reporter writes.
func TestSinkImplementable(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	sink := NewSink(func() (*Stream, error) {
		return NewStream(nopCloser{buf}), nil
	})

	stream, err := Open(sink)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(stream, "report body"); err != nil {
		t.Fatal(err)
	}

	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}

	if buf.String() != "report body" {
		t.Fatalf("captured %q", buf.String())
	}
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }
