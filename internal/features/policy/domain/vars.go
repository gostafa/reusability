// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"errors"
)

var (
	errNotFinite     = errors.New("must be a finite number")
	errPatternEmpty  = errors.New("pattern must be non-empty")
	errMinOutOfRange = errors.New("min must be in [0, 1]")
	errNegativeWrite = errors.New("negative write count")
)
