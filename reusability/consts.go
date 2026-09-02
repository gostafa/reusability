// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"github.com/gostafa/reusability/internal/shared/metrics"
)

const (
	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts every import, including external modules and
	// the standard library.
	DependencyScopeAll DependencyScope = "all"

	// FieldUsageDirect counts only fields a method body accesses directly.
	FieldUsageDirect FieldUsageMode = "direct"
	// FieldUsageTransitive additionally propagates field usage through calls
	// to sibling methods of the same type, to a fixpoint.
	FieldUsageTransitive FieldUsageMode = "transitive"

	// MetricReusability is the sole public metric this linter reports.
	MetricReusability MetricName = metrics.MetricReusability
	// SchemaVersion identifies the report schema produced by this build.
	SchemaVersion = "6"

	emptyString           = ""
	zero                  = 0
	defaultPackagePattern = "./..."

	errWrapQuoted = "%w: %q"
)
