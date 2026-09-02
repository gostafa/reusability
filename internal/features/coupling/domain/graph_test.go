package domain

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

func scopedFacts() *typefacts.ProjectFacts {
	return &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{
				ID: 0, Path: "example.com/m/a", InModule: true,
				Imports: []string{"example.com/m/b", "example.com/other/lib", "fmt"},
			},
			{ID: 1, Path: "example.com/m/b", InModule: true},
		},
	}
}

func TestBuildDependencyGraphScopes(t *testing.T) {
	facts := scopedFacts()

	project := BuildDependencyGraph(facts, ScopeProject)
	if got := project.Couplings[0]; got.Efferent != 1 || got.Afferent != 0 {
		t.Fatalf("project scope a = %+v", got)
	}

	if got := project.Couplings[1]; got.Afferent != 1 || got.Efferent != 0 {
		t.Fatalf("project scope b = %+v", got)
	}

	module := BuildDependencyGraph(facts, ScopeModule)
	if got := module.Couplings[0].Efferent; got != 1 {
		t.Fatalf("module scope Ce(a) = %d, want 1 (fmt and external excluded)", got)
	}

	all := BuildDependencyGraph(facts, ScopeAll)
	if got := all.Couplings[0].Efferent; got != 3 {
		t.Fatalf("all scope Ce(a) = %d, want 3", got)
	}

	// Afferent coupling is always measured within the analyzed set.
	if got := all.Couplings[1].Afferent; got != 1 {
		t.Fatalf("all scope Ca(b) = %d, want 1", got)
	}
}

func TestModuleScopeWithoutModuleInfo(t *testing.T) {
	facts := scopedFacts()
	facts.ModulePath = ""

	graph := BuildDependencyGraph(facts, ScopeModule)
	if got := graph.Couplings[0].Efferent; got != 1 {
		t.Fatalf("Ce(a) = %d, want 1 (degrades to project scope)", got)
	}
}

func TestCountTypes(t *testing.T) {
	facts := &typefacts.ProjectFacts{
		Packages: []typefacts.PackageFacts{{ID: 0, Path: "p", TypeIDs: []int{0, 1, 2}}},
		Types: []typefacts.TypeFacts{
			{ID: 0, Kind: typefacts.KindStruct},
			{ID: 1, Kind: typefacts.KindInterface},
			{ID: 2, Kind: typefacts.KindOther},
		},
	}

	counts := CountTypes(facts, &facts.Packages[0])
	if counts.Interfaces != 1 || counts.Total != 3 {
		t.Fatalf("counts = %+v", counts)
	}
}
