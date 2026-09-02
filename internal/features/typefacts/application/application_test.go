// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"fmt"
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
					"example.com/m/alpha.A",
					"example.com/m/missing.Gone",
				}},
				{Name: "A"},
			},
			Imports: []string{"fmt", "example.com/m/alpha", "fmt", "example.com/m/zeta"},
		},
		{
			Path:     "example.com/m/alpha",
			InModule: true,
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

	for i, pkg := range facts.Packages {
		if pkg.ID != i {
			t.Fatalf("package ID %d at index %d", pkg.ID, i)
		}
	}

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

	b := facts.Types[2]
	if len(b.ReferencedTypeIDs) != 1 || b.ReferencedTypeIDs[0] != 0 {
		t.Fatalf("B.ReferencedTypeIDs = %v, want [0]", b.ReferencedTypeIDs)
	}

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
	*outbound.FactOptions,
) (string, []domain.PackageExtract, error) {
	return "", nil, s.err
}

func TestCollectPropagatesLoadError(t *testing.T) {
	sentinel := errors.New("load failed")
	_, err := NewService(
		errSource{err: sentinel},
	).Collect(context.Background(), &outbound.FactOptions{})
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

func benchExtracts(pkgCount, typesPerPkg int) []domain.PackageExtract {
	pkgs := make([]domain.PackageExtract, pkgCount)
	for p := range pkgs {
		types := make([]domain.TypeExtract, typesPerPkg)
		for i := range types {
			types[i] = domain.TypeExtract{
				Name: fmt.Sprintf("Type%02d", i),
				ReferencedTypeKeys: []string{
					domain.TypeKey(fmt.Sprintf("example.com/m/pkg%d", (p+1)%pkgCount), "Type00"),
					domain.TypeKey(fmt.Sprintf("example.com/m/pkg%d", (p+2)%pkgCount), "Type01"),
				},
			}
		}

		pkgs[p] = domain.PackageExtract{
			Path:     fmt.Sprintf("example.com/m/pkg%d", p),
			InModule: true,
			Imports: []string{
				fmt.Sprintf("example.com/m/pkg%d", (p+1)%pkgCount),
				"fmt",
				"context",
			},
			Types: types,
		}
	}

	return pkgs
}

func BenchmarkAssemble(b *testing.B) {
	extracts := benchExtracts(60, 25)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {

		cp := make([]domain.PackageExtract, len(extracts))
		copy(cp, extracts)
		_ = Assemble("example.com/m", cp)
	}
}

type fakeSource struct{}

func (fakeSource) Load(
	context.Context,
	*outbound.FactOptions,
) (string, []domain.PackageExtract, error) {
	return "example.com/m", []domain.PackageExtract{
		{Path: "example.com/m/b", InModule: true, Types: []domain.TypeExtract{{Name: "B"}}},
		{Path: "example.com/m/a", InModule: true, Types: []domain.TypeExtract{{Name: "A"}}},
	}, nil
}

// Black-box: the service loads through the port and assembles deterministic,
// sorted facts with dense IDs.
func TestServiceCollect(t *testing.T) {
	t.Parallel()

	svc := NewService(fakeSource{})

	facts, err := svc.Collect(
		context.Background(),
		&outbound.FactOptions{Patterns: []string{"./..."}},
	)
	if err != nil {
		t.Fatal(err)
	}

	if facts.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", facts.ModulePath)
	}

	if len(facts.Packages) != 2 || facts.Packages[0].Path != "example.com/m/a" {
		t.Fatalf("packages not sorted by path: %+v", facts.Packages)
	}

	for i, p := range facts.Packages {
		if p.ID != i {
			t.Errorf("package %q ID = %d, want %d", p.Path, p.ID, i)
		}
	}
}
