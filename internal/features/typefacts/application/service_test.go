package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

func TestAssembleOrderingAndIDs(t *testing.T) {
	extracts := []domain.PackageExtract{
		{
			Path: "example.com/m/zeta",
			Types: []domain.TypeExtract{
				{Name: "B", ReferencedTypeKeys: []string{
					"example.com/m/alpha.A",
					"example.com/m/alpha.A",      // duplicate collapses
					"example.com/m/missing.Gone", // outside the analyzed set
				}},
				{Name: "A"},
			},
			Imports: []string{"fmt", "example.com/m/alpha", "fmt", "example.com/m/zeta"},
		},
		{
			Path:       "example.com/m/alpha",
			InModule:   true,
			VarCount:   4,
			ConstCount: 5,
			Variables: []domain.DeclarationFacts{{
				Name:     "AlphaVar",
				Exported: true,
				Pos:      domain.Position{File: "alpha/a.go", Line: 3, Column: 5},
			}},
			Constants: []domain.DeclarationFacts{{
				Name: "alphaConst",
				Pos:  domain.Position{File: "alpha/a.go", Line: 4, Column: 7},
			}},
			Functions: []domain.FunctionFacts{{
				Name:     "AlphaFunc",
				Exported: true,
				Pos:      domain.Position{File: "alpha/a.go", Line: 6, Column: 1},
				Lines:    3,
				Branches: domain.BranchStats{Ifs: 1},
			}},
			Types: []domain.TypeExtract{{
				Name: "A",
				Fields: []domain.FieldFacts{{
					Name:     "Field",
					Exported: true,
				}},
				Methods: []domain.MethodFacts{{
					Name:     "Do",
					Exported: true,
					Pos:      domain.Position{File: "alpha/a.go", Line: 10, Column: 1},
					Lines:    1,
					Branches: domain.BranchStats{LogicalOps: 1},
				}},
			}},
		},
	}

	facts := Assemble("example.com/m", extracts)

	if facts.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", facts.ModulePath)
	}

	if len(facts.Packages) != 2 || facts.Packages[0].Path != "example.com/m/alpha" {
		t.Fatalf("packages not sorted by path: %+v", facts.Packages)
	}
	if facts.Packages[0].VarCount != 4 || facts.Packages[0].ConstCount != 5 {
		t.Fatalf("alpha counts = vars %d consts %d, want 4 and 5",
			facts.Packages[0].VarCount, facts.Packages[0].ConstCount)
	}
	if facts.Packages[0].Variables[0].Name != "AlphaVar" ||
		facts.Packages[0].Functions[0].Lines != 3 {
		t.Fatalf("alpha declaration details = %+v", facts.Packages[0])
	}

	for i, pkg := range facts.Packages {
		if pkg.ID != i {
			t.Fatalf("package ID %d at index %d", pkg.ID, i)
		}
	}

	// Types globally sorted by (package path, name) with dense IDs.
	wantNames := []string{"A", "A", "B"}
	for i, typ := range facts.Types {
		if typ.ID != i || typ.Name != wantNames[i] {
			t.Fatalf("types[%d] = {ID:%d Name:%q}, want {ID:%d Name:%q}",
				i, typ.ID, typ.Name, i, wantNames[i])
		}
	}
	if facts.Types[0].Fields[0].Name != "Field" ||
		facts.Types[0].Methods[0].Lines != 1 {
		t.Fatalf("type details were not preserved: %+v", facts.Types[0])
	}

	// zeta.B (last type) references alpha.A (ID 0); the missing key dropped.
	b := facts.Types[2]
	if len(b.ReferencedTypeIDs) != 1 || b.ReferencedTypeIDs[0] != 0 {
		t.Fatalf("B.ReferencedTypeIDs = %v, want [0]", b.ReferencedTypeIDs)
	}

	// Imports deduplicated, sorted, self-import removed.
	zeta := facts.Packages[1]
	if len(zeta.Imports) != 2 || zeta.Imports[0] != "example.com/m/alpha" ||
		zeta.Imports[1] != "fmt" {
		t.Fatalf("zeta.Imports = %v", zeta.Imports)
	}

	if len(zeta.TypeIDs) != 2 || facts.Types[zeta.TypeIDs[0]].Name != "A" {
		t.Fatalf("zeta.TypeIDs = %v", zeta.TypeIDs)
	}
}

type errSource struct{ err error }

func (s errSource) Load(
	context.Context,
	outbound.FactOptions,
) (string, []domain.PackageExtract, error) {
	return "", nil, s.err
}

func TestCollectPropagatesLoadError(t *testing.T) {
	sentinel := errors.New("load failed")
	_, err := NewService(
		errSource{err: sentinel},
	).Collect(context.Background(), outbound.FactOptions{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Collect error = %v, want sentinel", err)
	}
}

func TestResolveKeysAllMissing(t *testing.T) {
	facts := Assemble("example.com/m", []domain.PackageExtract{
		{
			Path: "example.com/m/p",
			Types: []domain.TypeExtract{
				{Name: "T", ReferencedTypeKeys: []string{"example.com/m/other.U"}},
			},
		},
	})
	if ids := facts.Types[0].ReferencedTypeIDs; ids != nil {
		t.Fatalf("ReferencedTypeIDs = %v, want nil", ids)
	}
}

func TestSortedUniqueSelfOnly(t *testing.T) {
	facts := Assemble("example.com/m", []domain.PackageExtract{{
		Path:    "example.com/m/p",
		Imports: []string{"example.com/m/p"},
		Types:   []domain.TypeExtract{{Name: "T"}},
	}})
	if imports := facts.Packages[0].Imports; imports != nil {
		t.Fatalf("Imports = %v, want nil", imports)
	}
}
