// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

var errStdoutUnavailable = errors.New("stdout is unavailable")

// Open creates (or truncates) the sink's destination file.
func (sink FileSink) Open() (*outbound.Stream, error) {
	file, createErr := os.Create(sink.Path)
	if createErr != nil {
		return nil, fmt.Errorf("create report file: %w", createErr)
	}

	return outbound.NewStream(file), nil
}

// Open returns a buffered writer for standard output.
func (StdoutSink) Open() (*outbound.Stream, error) {
	if os.Stdout == nil {
		return nil, errStdoutUnavailable
	}

	return outbound.NewStream(stdoutStream{w: bufio.NewWriter(os.Stdout)}), nil
}

// Close flushes buffered output without closing stdout.
func (stream stdoutStream) Close() error {
	flushErr := stream.w.Flush()
	if flushErr != nil {
		return fmt.Errorf("stdout flush: %w", flushErr)
	}

	return nil
}

// Write buffers payload for standard output.
func (stream stdoutStream) Write(payload []byte) (int, error) {
	count, writeErr := stream.w.Write(payload)
	if writeErr != nil {
		return count, fmt.Errorf("stdout write: %w", writeErr)
	}

	return count, nil
}
