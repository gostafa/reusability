// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"errors"
)

var (
	errEmptyPattern       = errors.New("empty pattern in rule spec")
	errExpectedPatternMin = errors.New("expected pattern:min")
	errNoPolicyRules      = errors.New(
		"no policy rules configured; pass at least one -rule=pattern:min with -check",
	)
	errShortWrite = errors.New("fprint: short write")
)
