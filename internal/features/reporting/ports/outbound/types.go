// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"io"
)

type (
	// Sink opens a concrete stream for rendered report bytes.
	Sink = struct {
		open func() (*Stream, error)
	}

	// Stream is a concrete closable writer for report output.
	Stream struct {
		writeCloser io.WriteCloser
	}

	writer interface {
		Write(payload []byte) (int, error)
	}

	closer interface {
		Close() error
	}

	destWriter interface {
		destWrite(payload []byte) (int, error)
	}
)
