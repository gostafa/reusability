// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

type (
	typeBuild struct {
		idByKey map[string]int
		extract domain.TypeExtract
		id      int
		pkgID   int
	}

	// Collector loads and assembles project type facts.
	Collector interface {
		Collect(ctx context.Context, opts *outbound.FactOptions) (domain.ProjectFacts, error)
	}

	// Service is the application service backed by a FactSource.
	Service struct {
		source outbound.FactSource
	}
)
