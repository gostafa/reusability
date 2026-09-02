// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"errors"
)

var (
	errFprintShortWrite  = errors.New("fprint: short write")
	errPrintlnNegCount   = errors.New("println: negative write count")
	errBuilderShortWrite = errors.New("builder: short write")
)
