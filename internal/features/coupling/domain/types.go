// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (
	// Scope selects which imports count toward efferent coupling.
	Scope string

	// Coupling holds afferent and efferent dependency counts for one package.
	Coupling struct {
		// Afferent counts analyzed packages importing this package (Ca).
		Afferent int
		// Efferent counts this package's in-scope imports (Ce).
		Efferent int
	}

	// CouplingGraph looks up coupling counts by package ID.
	CouplingGraph interface {
		Coupling(packageID int) Coupling
	}

	// DependencyGraph stores per-package coupling derived from facts.
	DependencyGraph struct {
		// Couplings holds each package's dependency counts, indexed by ID.
		Couplings []Coupling
	}

	// TypeCounts tallies interface and total named types in a package.
	TypeCounts struct {
		// Interfaces counts the package's named interface types.
		Interfaces int
		// Total counts all of the package's analyzed named types.
		Total int
	}

	scopeCheck struct {
		analyzed   map[string]int
		path       string
		modulePath string
		scope      Scope
	}
)
