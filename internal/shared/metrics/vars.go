// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

import (
	"errors"
)

var (
	_ Validator = (*ReusabilityWeights)(nil)

	errWeightsSumZero = errors.New("reusability weights sum to zero and cannot be normalized")
	errNegativeWeight = errors.New("reusability weight is negative")
)
