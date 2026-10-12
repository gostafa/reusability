// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"errors"

	"github.com/gostafa/reusability/internal/shared/zerovalue"
)

var (
	errNotFinite     = errors.New("must be a finite number")
	errPatternEmpty  = errors.New("pattern must be non-empty")
	errMinOutOfRange = errors.New("min must be in [0, 1]")

	_ PathMatcher = zerovalue.Of[Rule]()
	_ RuleGate    = zerovalue.Of[Rule]()
)
