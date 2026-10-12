// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
)

// OpenFile creates (or truncates) the sink's destination file.
func OpenFile(sink FileSink) (*outbound.Stream, error) {
	file, err := os.Create(sink.Path)
	if err != nil {
		return nil, fmt.Errorf("create report file: %w", err)
	}

	return outbound.NewStream(file), nil
}

// OpenStdout returns a buffered writer for standard output.
func OpenStdout() (*outbound.Stream, error) {
	if os.Stdout == nil {
		return nil, errStdoutUnavailable
	}

	stream := stdoutStream{writer: bufio.NewWriter(os.Stdout)}

	return outbound.NewStream(stream), nil
}

// Close flushes buffered output without closing stdout.
func (stream stdoutStream) Close() error {
	err := stream.writer.Flush()
	if err != nil {
		return fmt.Errorf("flush stdout: %w", err)
	}

	return nil
}

// Write buffers payload for standard output.
func (stream stdoutStream) Write(payload []byte) (int, error) {
	written, err := stream.writer.Write(payload)
	if err != nil {
		return written, fmt.Errorf(
			"write stdout (%d buffered, %d available): %w",
			stream.buffered(), stream.available(), err,
		)
	}

	return written, nil
}

func (stream stdoutStream) buffered() int {
	return stream.writer.Buffered()
}

func (stream stdoutStream) available() int {
	return stream.writer.Available()
}
