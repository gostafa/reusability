package application

import (
	"context"
	"sort"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

// Collector is the type-facts application boundary consumed by analysis
// pipelines. Implementations load and assemble one deterministic project view.
type Collector interface {
	Collect(context.Context, outbound.FactOptions) (domain.ProjectFacts, error)
}

// Service produces ProjectFacts through the outbound fact source.
type Service struct {
	source outbound.FactSource
}

var _ Collector = (*Service)(nil)

// NewService returns a Service backed by the given fact source.
func NewService(source outbound.FactSource) *Service {
	return &Service{source: source}
}

// Collect loads the project once and returns its assembled facts.
func (s *Service) Collect(
	ctx context.Context,
	opts outbound.FactOptions,
) (domain.ProjectFacts, error) {
	modulePath, extracts, err := s.source.Load(ctx, opts)
	if err != nil {
		return domain.ProjectFacts{}, err
	}

	return Assemble(modulePath, extracts), nil
}

// Assemble sorts the extracts, assigns dense numeric IDs, and resolves
// referenced-type keys to type IDs. Ordering is fully deterministic:
// packages by import path, types by (package path, name); field and method
// order is preserved from the extraction contract.
func Assemble(modulePath string, extracts []domain.PackageExtract) domain.ProjectFacts {
	sort.Slice(extracts, func(i, j int) bool { return extracts[i].Path < extracts[j].Path })

	for i := range extracts {
		types := extracts[i].Types
		sort.Slice(types, func(a, b int) bool { return types[a].Name < types[b].Name })
	}

	totalTypes := 0
	for i := range extracts {
		totalTypes += len(extracts[i].Types)
	}

	facts := domain.ProjectFacts{
		ModulePath: modulePath,
		Packages:   make([]domain.PackageFacts, 0, len(extracts)),
		Types:      make([]domain.TypeFacts, 0, totalTypes),
	}

	idByKey := make(map[string]int, totalTypes)
	nextID := 0

	for _, extract := range extracts {
		for _, t := range extract.Types {
			idByKey[domain.TypeKey(extract.Path, t.Name)] = nextID
			nextID++
		}
	}

	typeID := 0

	for pkgID, extract := range extracts {
		pkg := domain.PackageFacts{
			ID:       pkgID,
			Path:     extract.Path,
			InModule: extract.InModule,
			Imports:  sortedUnique(extract.Imports, extract.Path),
			TypeIDs:  make([]int, 0, len(extract.Types)),
		}
		for _, t := range extract.Types {
			// IDs were assigned in this same iteration order above, so a
			// running counter matches idByKey without recomputing the key.
			id := typeID
			typeID++

			pkg.TypeIDs = append(pkg.TypeIDs, id)
			facts.Types = append(facts.Types, domain.TypeFacts{
				ID:                        id,
				PackageID:                 pkgID,
				Name:                      t.Name,
				Exported:                  t.Exported,
				Kind:                      t.Kind,
				Pos:                       t.Pos,
				Fields:                    t.Fields,
				Methods:                   t.Methods,
				ReferencedTypeIDs:         resolveKeys(t.ReferencedTypeKeys, idByKey),
				ExportedMembers:           t.ExportedMembers,
				DocumentedExportedMembers: t.DocumentedExportedMembers,
			})
		}

		facts.Packages = append(facts.Packages, pkg)
	}

	return facts
}

// resolveKeys maps referenced-type keys to sorted, deduplicated type IDs.
// Keys outside the analyzed set (already filtered by the adapter) are
// dropped defensively.
func resolveKeys(keys []string, idByKey map[string]int) []int {
	if len(keys) == 0 {
		return nil
	}

	ids := make([]int, 0, len(keys))
	for _, key := range keys {
		if id, ok := idByKey[key]; ok {
			ids = append(ids, id)
		}
	}

	sort.Ints(ids)

	ids = uniqueInts(ids)
	if len(ids) == 0 {
		return nil
	}

	return ids
}

func uniqueInts(sorted []int) []int {
	out := sorted[:0]
	for i, v := range sorted {
		if i == 0 || v != sorted[i-1] {
			out = append(out, v)
		}
	}

	return out
}

// sortedUnique sorts and deduplicates import paths and removes self-imports.
func sortedUnique(imports []string, self string) []string {
	if len(imports) == 0 {
		return nil
	}

	out := make([]string, 0, len(imports))
	for _, path := range imports {
		if path != self {
			out = append(out, path)
		}
	}

	sort.Strings(out)

	dedup := out[:0]
	for i, path := range out {
		if i == 0 || path != out[i-1] {
			dedup = append(dedup, path)
		}
	}

	if len(dedup) == 0 {
		return nil
	}

	return dedup
}
