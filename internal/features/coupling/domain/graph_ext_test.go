package domain_test

import (
	"testing"

	architecture "github.com/gostafa/reusability/internal/features/coupling/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// Black-box: coupling counts and type tallies from the exported entry points.
func TestGraphAndCounts(t *testing.T) {
	t.Parallel()

	facts := &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{
				ID: 0, Path: "example.com/m/a", InModule: true,
				Imports: []string{"example.com/m/b"}, TypeIDs: []int{0},
			},
			{ID: 1, Path: "example.com/m/b", InModule: true, TypeIDs: []int{1, 2}},
		},
		Types: []typefacts.TypeFacts{
			{ID: 0, PackageID: 0, Kind: typefacts.KindStruct},
			{ID: 1, PackageID: 1, Kind: typefacts.KindInterface},
			{ID: 2, PackageID: 1, Kind: typefacts.KindStruct},
		},
	}

	g := architecture.BuildDependencyGraph(facts, architecture.Scope("project"))
	if g.Couplings[0].Efferent != 1 {
		t.Errorf("a efferent = %d, want 1", g.Couplings[0].Efferent)
	}

	if g.Couplings[1].Afferent != 1 {
		t.Errorf("b afferent = %d, want 1", g.Couplings[1].Afferent)
	}

	counts := architecture.CountTypes(facts, &facts.Packages[1])
	if counts.Total != 2 || counts.Interfaces != 1 {
		t.Fatalf("b counts = %+v, want {Total:2 Interfaces:1}", counts)
	}
}
