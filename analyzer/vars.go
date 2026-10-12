// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"errors"
)

var (
	errRuleMinRequired        = errors.New("min is required")
	errInvalidDependencyScope = errors.New(
		"invalid dependency-scope (want project, module, or all)",
	)
	errInvalidFieldUsage = errors.New("invalid field-usage (want direct or transitive)")
)
