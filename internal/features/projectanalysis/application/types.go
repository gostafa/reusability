// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
)

type (
	// Pipeline orchestrates fact collection and metric computation.
	Pipeline struct {
		analyze func(context.Context, *inbound.Options) (inbound.Result, error)
	}
)
