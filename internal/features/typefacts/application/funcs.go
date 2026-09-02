// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

// Assemble sorts the extracts, assigns dense numeric IDs, and resolves
// referenced-type keys to type IDs. Ordering is fully deterministic:
// packages by import path, types by (package path, name); field and method
// order is preserved from the extraction contract.
func Assemble(modulePath string, extracts []domain.PackageExtract) domain.ProjectFacts {
	sortExtracts(extracts)

	facts := domain.ProjectFacts{
		ModulePath: modulePath,
		Packages:   make([]domain.PackageFacts, zero, len(extracts)),
		Types:      make([]domain.TypeFacts, zero, countTypes(extracts)),
	}

	appendPackages(&facts, extracts, indexTypeKeys(extracts))

	return facts
}

// NewService returns a Service backed by the given fact source.
func NewService(source outbound.FactSource) *Service {
	return &Service{source: source}
}

// Collect loads the project once and returns its assembled facts.
func (svc *Service) Collect(
	ctx context.Context,
	opts *outbound.FactOptions,
) (domain.ProjectFacts, error) {
	// Load package extracts from the fact source then assemble project facts.
	modulePath, extracts, err := svc.source.Load(ctx, opts)
	if err != nil {
		return domain.ProjectFacts{}, fmt.Errorf("Collect: %w", err)
	}

	return Assemble(modulePath, extracts), nil
}

func appendPackages(
	facts *domain.ProjectFacts,
	extracts []domain.PackageExtract,
	idByKey map[string]int,
) {
	// Append packages in extract order while assigning dense type IDs.
	typeID := zero

	for pkgID := range extracts {
		pkg, next := packageFacts(&packageBuild{
			pkgID: pkgID, extract: extracts[pkgID], typeID: typeID, idByKey: idByKey, facts: facts,
		})

		typeID = next

		facts.Packages = append(facts.Packages, pkg)
	}
}

func countTypes(extracts []domain.PackageExtract) int {
	total := zero

	for index := range extracts {
		total += len(extracts[index].Types)
	}

	return total
}

func indexTypeKeys(extracts []domain.PackageExtract) map[string]int {
	idByKey := make(map[string]int, countTypes(extracts))
	nextID := zero

	for index := range extracts {
		types := extracts[index].Types

		for typeIndex := range types {
			idByKey[domain.TypeKey(extracts[index].Path, types[typeIndex].Name)] = nextID
			nextID++
		}
	}

	return idByKey
}

func packageFacts(build *packageBuild) (pkg domain.PackageFacts, nextTypeID int) {
	pkg = domain.PackageFacts{
		ID:       build.pkgID,
		Path:     build.extract.Path,
		InModule: build.extract.InModule,
		Imports:  sortedUnique(build.extract.Imports, build.extract.Path),
		TypeIDs:  make([]int, zero, len(build.extract.Types)),
	}

	for index := range build.extract.Types {
		pkg.TypeIDs = append(pkg.TypeIDs, build.typeID)
		build.facts.Types = append(build.facts.Types, typeFacts(&typeBuild{
			id:      build.typeID,
			pkgID:   build.pkgID,
			extract: build.extract.Types[index],
			idByKey: build.idByKey,
		}))
		build.typeID++
	}

	return pkg, build.typeID
}

func resolveKeys(keys []string, idByKey map[string]int) []int {
	ids := mapKeys(keys, idByKey)

	if len(ids) == zero {
		return nil
	}

	slices.Sort(ids)

	return orNil(uniqueInts(ids))
}

func mapKeys(keys []string, idByKey map[string]int) []int {
	ids := make([]int, zero, len(keys))

	for index := range keys {
		if id, ok := idByKey[keys[index]]; ok {
			ids = append(ids, id)
		}
	}

	return ids
}

func orNil(ids []int) []int {
	if len(ids) == zero {
		return nil
	}

	return ids
}

func sortExtracts(extracts []domain.PackageExtract) {
	slices.SortFunc(extracts, func(left, right domain.PackageExtract) int {
		return cmp.Compare(left.Path, right.Path)
	})

	for index := range extracts {
		slices.SortFunc(extracts[index].Types, func(left, right domain.TypeExtract) int {
			return cmp.Compare(left.Name, right.Name)
		})
	}
}

func sortedUnique(imports []string, self string) []string {
	out := filterSelf(imports, self)

	if len(out) == zero {
		return nil
	}

	slices.Sort(out)

	return dedupSorted(out)
}

func dedupSorted(sorted []string) []string {
	dedup := sorted[:zero]

	for index := range sorted {
		path := sorted[index]

		if index == zero || path != sorted[index-one] {
			dedup = append(dedup, path)
		}
	}

	if len(dedup) == zero {
		return nil
	}

	return dedup
}

func filterSelf(imports []string, self string) []string {
	if len(imports) == zero {
		return nil
	}

	out := make([]string, zero, len(imports))

	for index := range imports {
		if imports[index] != self {
			out = append(out, imports[index])
		}
	}

	return out
}

func typeFacts(build *typeBuild) domain.TypeFacts {
	return domain.TypeFacts{
		ID:                        build.id,
		PackageID:                 build.pkgID,
		Name:                      build.extract.Name,
		Exported:                  build.extract.Exported,
		Kind:                      build.extract.Kind,
		Pos:                       build.extract.Pos,
		Fields:                    build.extract.Fields,
		Methods:                   build.extract.Methods,
		ReferencedTypeIDs:         resolveKeys(build.extract.ReferencedTypeKeys, build.idByKey),
		ExportedMembers:           build.extract.ExportedMembers,
		DocumentedExportedMembers: build.extract.DocumentedExportedMembers,
	}
}

func uniqueInts(sorted []int) []int {
	out := sorted[:zero]

	for index := range sorted {
		value := sorted[index]

		if index == zero || value != sorted[index-one] {
			out = append(out, value)
		}
	}

	return out
}
