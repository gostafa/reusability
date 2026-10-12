// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"errors"
)

var (
	// ErrConfig identifies invalid analysis configuration.
	ErrConfig = errors.New("validate config")
	// ErrAnalysis identifies an analysis pipeline failure.
	ErrAnalysis               = errors.New("analyze")
	errEmptyPackagePattern    = errors.New("empty package pattern")
	errInvalidDependencyScope = errors.New(
		"invalid dependency scope (want project, module, or all)",
	)
	errInvalidFieldUsage = errors.New(
		"invalid field usage mode (want direct or transitive)",
	)
)
