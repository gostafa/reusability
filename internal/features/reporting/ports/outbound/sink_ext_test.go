package outbound_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

// memSink is an external adapter capturing report output in a buffer.
type memSink struct{ buf *bytes.Buffer }

func (m memSink) Open() (io.WriteCloser, error) { return nopCloser{m.buf}, nil }

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

// Black-box: an external Sink can capture what the reporter writes.
func TestSinkImplementable(t *testing.T) {
	t.Parallel()

	sink := memSink{buf: &bytes.Buffer{}}

	var s outbound.Sink = sink

	w, err := s.Open()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(w, "report body"); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	if sink.buf.String() != "report body" {
		t.Fatalf("captured %q", sink.buf.String())
	}
}
