package domain

import (
	"strings"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// Scope selects which import edges count toward efferent coupling.
type Scope string

const (
	// ScopeProject counts only imports of other analyzed packages.
	ScopeProject Scope = "project"
	// ScopeModule counts imports of packages in the main module. Without
	// module information it degrades to ScopeProject.
	ScopeModule Scope = "module"
	// ScopeAll counts every import.
	ScopeAll Scope = "all"
)

// Coupling carries one package's dependency counts. Afferent coupling is
// always measured within the analyzed set (importers outside it are not
// observable); efferent coupling honors the scope.
type Coupling struct {
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int
}

// CouplingGraph exposes read-only package coupling by package ID. Consumers
// depend on this narrow view instead of the graph's concrete storage.
type CouplingGraph interface {
	Coupling(packageID int) Coupling
}

// DependencyGraph is the analyzed packages' coupling, indexed by package ID.
// It is built once per analysis run.
type DependencyGraph struct {
	// Couplings holds each package's dependency counts, indexed by ID.
	Couplings []Coupling
}

// Coupling returns the dependency counts for packageID.
func (g DependencyGraph) Coupling(packageID int) Coupling {
	return g.Couplings[packageID]
}

var _ CouplingGraph = DependencyGraph{}

// BuildDependencyGraph derives the dependency graph from the package facts.
// Import lists are already deduplicated and free of self-edges.
func BuildDependencyGraph(facts *typefacts.ProjectFacts, scope Scope) DependencyGraph {
	analyzed := make(map[string]int, len(facts.Packages))
	for i := range facts.Packages {
		analyzed[facts.Packages[i].Path] = i
	}

	if facts.ModulePath == "" && scope == ScopeModule {
		scope = ScopeProject
	}

	graph := DependencyGraph{Couplings: make([]Coupling, len(facts.Packages))}
	for i := range facts.Packages {
		for _, path := range facts.Packages[i].Imports {
			if target, ok := analyzed[path]; ok {
				graph.Couplings[target].Afferent++
			}

			if inScope(path, scope, facts.ModulePath, analyzed) {
				graph.Couplings[i].Efferent++
			}
		}
	}

	return graph
}

func inScope(path string, scope Scope, modulePath string, analyzed map[string]int) bool {
	switch scope {
	case ScopeAll:
		return true
	case ScopeModule:
		return path == modulePath || strings.HasPrefix(path, modulePath+"/")
	default: // ScopeProject
		_, ok := analyzed[path]

		return ok
	}
}

// TypeCounts summarizes one package's named types for Abstractness. Aliases
// never reach the fact model, so every analyzed type is relevant.
type TypeCounts struct {
	// Interfaces counts the package's named interface types.
	Interfaces int
	// Total counts all of the package's analyzed named types.
	Total int
}

// CountTypes tallies a package's interface and total named types.
func CountTypes(facts *typefacts.ProjectFacts, pkg *typefacts.PackageFacts) TypeCounts {
	var counts TypeCounts
	for _, id := range pkg.TypeIDs {
		counts.Total++
		if facts.Types[id].Kind == typefacts.KindInterface {
			counts.Interfaces++
		}
	}

	return counts
}
