// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
)

type (
	// Pipeline orchestrates fact collection and metric computation.
	Pipeline struct {
		facts typefacts.Collector
	}
)
