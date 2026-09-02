// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// BuildDependencyGraph derives the dependency graph from the package facts.
// Import lists are already deduplicated and free of self-edges.
func BuildDependencyGraph(facts *typefacts.ProjectFacts, scope Scope) DependencyGraph {
	analyzed := indexPackages(facts)

	scope = normalizeScope(facts.ModulePath, scope)

	graph := DependencyGraph{Couplings: make([]Coupling, len(facts.Packages))}

	for index := range facts.Packages {
		addPackageEdges(&edgeJob{
			graph: &graph, facts: facts, pkgID: index, scope: scope, analyzed: analyzed,
		})
	}

	return graph
}

// CountTypes tallies a package's interface and total named types.
func CountTypes(facts *typefacts.ProjectFacts, pkg *typefacts.PackageFacts) TypeCounts {
	var counts TypeCounts

	for index := range pkg.TypeIDs {
		counts.Total++

		if facts.Types[pkg.TypeIDs[index]].Kind == typefacts.KindInterface {
			counts.Interfaces++
		}
	}

	return counts
}

// Coupling returns the dependency counts for packageID.
func (graph DependencyGraph) Coupling(packageID int) Coupling {
	return graph.Couplings[packageID]
}

func addPackageEdges(job *edgeJob) {
	imports := job.facts.Packages[job.pkgID].Imports

	for index := range imports {
		recordEdge(job, imports[index])
	}
}

func recordEdge(job *edgeJob, path string) {
	recordAfferent(job, path)
	recordEfferent(job, path)
}

func recordAfferent(job *edgeJob, path string) {
	if target, ok := job.analyzed[path]; ok {
		job.graph.Couplings[target].Afferent++
	}
}

func recordEfferent(job *edgeJob, path string) {
	check := &scopeCheck{
		path: path, scope: job.scope, modulePath: job.facts.ModulePath, analyzed: job.analyzed,
	}

	if inScope(check) {
		job.graph.Couplings[job.pkgID].Efferent++
	}
}

func indexPackages(facts *typefacts.ProjectFacts) map[string]int {
	analyzed := make(map[string]int, len(facts.Packages))

	for index := range facts.Packages {
		analyzed[facts.Packages[index].Path] = index
	}

	return analyzed
}

func inScope(check *scopeCheck) bool {
	if check.scope == ScopeAll {
		return true
	}

	if check.scope == ScopeModule {
		return isModuleImport(check)
	}

	return isAnalyzedImport(check)
}

func isAnalyzedImport(check *scopeCheck) bool {
	packageID, exists := check.analyzed[check.path]

	return exists && packageID >= zero
}

func isModuleImport(check *scopeCheck) bool {
	return check.path == check.modulePath ||
		strings.HasPrefix(check.path, check.modulePath+"/")
}

func normalizeScope(modulePath string, scope Scope) Scope {
	if modulePath == emptyPath && scope == ScopeModule {
		return ScopeProject
	}

	return scope
}
