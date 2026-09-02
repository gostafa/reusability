package sinks

import (
	"bufio"
	"io"
	"os"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

// Compile-time proof that both adapters satisfy the reporting outbound port.
var (
	_ outbound.Sink = StdoutSink{}
	_ outbound.Sink = FileSink{}
)

// StdoutSink writes the report to standard output. Closing the returned
// stream flushes it — stdout itself stays open, the process owns it.
type StdoutSink struct{}

// Open returns a buffered stdout stream.
func (StdoutSink) Open() (io.WriteCloser, error) {
	return stdoutStream{w: bufio.NewWriter(os.Stdout)}, nil
}

// stdoutStream buffers report output to standard output.
type stdoutStream struct {
	w *bufio.Writer
}

// Write buffers p for standard output.
func (s stdoutStream) Write(p []byte) (int, error) { return s.w.Write(p) }

// Close flushes buffered output without closing stdout.
func (s stdoutStream) Close() error { return s.w.Flush() }

// FileSink writes the report to a file, truncating any existing content.
type FileSink struct {
	// Path is the destination file path passed to os.Create.
	Path string
}

// Open creates the destination file.
func (s FileSink) Open() (io.WriteCloser, error) {
	return os.Create(s.Path)
}
