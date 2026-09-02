package application

import (
	"fmt"
	"testing"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
)

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
		// Assemble mutates (sorts) its input in place; copy per iteration so
		// each run sees identical work.
		cp := make([]domain.PackageExtract, len(extracts))
		copy(cp, extracts)
		_ = Assemble("example.com/m", cp)
	}
}
