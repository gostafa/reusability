// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"bufio"
)

type (
	// StdoutSink writes report output to standard output.
	StdoutSink struct{}

	stdoutStream struct {
		w *bufio.Writer
	}

	// FileSink writes report output to the file at Path.
	FileSink struct {
		// Path is the destination file path passed to os.Create.
		Path string
	}
)
