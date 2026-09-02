// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"fmt"
	"io"
)

// NewSink builds a Sink from an open function.
func NewSink(open func() (*Stream, error)) Sink {
	return Sink{open: open}
}

// Open returns the stream; the caller closes it when rendering is done.
func (sink Sink) Open() (*Stream, error) {
	stream, err := sink.open()
	if err != nil {
		return nil, fmt.Errorf("open sink: %w", err)
	}

	return stream, nil
}

// NewStream wraps a WriteCloser as a concrete Stream.
func NewStream(writeCloser io.WriteCloser) *Stream {
	return &Stream{writeCloser: writeCloser}
}

// Write writes p to the underlying writer.
func (stream *Stream) Write(payload []byte) (int, error) {
	written, err := stream.writeCloser.Write(payload)
	if err != nil {
		return written, fmt.Errorf("stream write: %w", err)
	}

	return written, nil
}

// Close closes the underlying closer.
func (stream *Stream) Close() error {
	err := stream.writeCloser.Close()
	if err != nil {
		return fmt.Errorf("stream close: %w", err)
	}

	return nil
}
