// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

type (
	packageBuild = struct {
		facts   *domain.ProjectFacts
		extract *domain.PackageExtract
		idByKey map[string]int
		pkgID   int
		typeID  int
	}

	typeBuild = struct {
		idByKey map[string]int
		extract domain.TypeExtract
		id      int
		pkgID   int
	}

	// Service is the application service backed by a FactSource.
	Service = func(context.Context, *outbound.FactOptions) (domain.ProjectFacts, error)

	factSourceFunc = func(context.Context, *outbound.FactOptions) (string, []domain.PackageExtract, error)
)
