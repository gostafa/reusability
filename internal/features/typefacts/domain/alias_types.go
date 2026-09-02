// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	factmodel "github.com/gostafa/reusability/internal/features/typefacts/domain/model"
)

type (
	// FieldFacts describes one struct field.
	FieldFacts = factmodel.FieldFacts
	// DeclarationFacts is shared name/position/export metadata.
	DeclarationFacts = factmodel.DeclarationFacts
	// FunctionFacts describes a package-level function.
	FunctionFacts = factmodel.FunctionFacts
	// MethodFacts describes one method of a named type.
	MethodFacts = factmodel.MethodFacts
	// BranchStats counts control-flow constructs for cyclomatic complexity.
	BranchStats = factmodel.BranchStats
)
