// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// ScopeProject counts only imports of other analyzed packages.
	ScopeProject Scope = "project"
	// ScopeModule counts imports of packages in the main module. Without
	// module information it degrades to ScopeProject.
	ScopeModule Scope = "module"
	// ScopeAll counts every import.
	ScopeAll Scope = "all"

	emptyPath = ""
	zero      = 0
)
