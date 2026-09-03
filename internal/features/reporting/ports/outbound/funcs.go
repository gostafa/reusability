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

// NewStream wraps a WriteCloser as a concrete Stream.
func NewStream(writeCloser io.WriteCloser) *Stream {
	return &Stream{writeCloser: writeCloser}
}

// Close closes the underlying closer.
func (stream *Stream) Close() error {
	err := closeOut(stream.writeCloser)
	if err != nil {
		return fmt.Errorf("stream close: %w", err)
	}

	return nil
}

func (stream *Stream) destWrite(payload []byte) (int, error) {
	written, err := writeBytes(stream.writeCloser, payload)
	if err != nil {
		return written, fmt.Errorf("dest write: %w", err)
	}

	return written, nil
}

// Open returns the stream; the caller closes it when rendering is done.
// Open returns the stream; the caller closes it when rendering is done.
func Open(sink Sink) (*Stream, error) {
	stream, err := sink.open()
	if err != nil {
		return nil, fmt.Errorf("open sink: %w", err)
	}

	return stream, nil
}

func closeOut(source closer) error {
	err := source.Close()
	if err != nil {
		return fmt.Errorf("close out: %w", err)
	}

	return nil
}

// Write writes p to the underlying writer.
func (stream *Stream) Write(payload []byte) (int, error) {
	written, err := destWriteOf(stream, payload)
	if err != nil {
		return written, fmt.Errorf("stream write: %w", err)
	}

	return written, nil
}

func destWriteOf(source destWriter, payload []byte) (int, error) {
	written, err := source.destWrite(payload)
	if err != nil {
		return written, fmt.Errorf("dest write of: %w", err)
	}

	return written, nil
}

func writeBytes(output writer, payload []byte) (int, error) {
	written, err := output.Write(payload)
	if err != nil {
		return written, fmt.Errorf("write bytes: %w", err)
	}

	return written, nil
}
