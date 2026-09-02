// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"context"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
)

type (
	// FactOptions configures one project load.
	FactOptions struct {
		Directory        string
		Patterns         []string
		BuildTags        []string
		Workers          int
		IncludeTests     bool
		IncludeGenerated bool
		ContinueOnError  bool
	}

	// FactSource loads package extracts for analysis.
	FactSource interface {
		// Load returns the main module path (empty when unknown) and one
		// PackageExtract per analyzed package, honoring the ordering contract
		// documented on domain.PackageExtract.
		Load(
			ctx context.Context,
			opts *FactOptions,
		) (modulePath string, packages []domain.PackageExtract, err error)
	}
)
