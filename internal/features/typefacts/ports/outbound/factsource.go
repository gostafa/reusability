package outbound

import (
	"context"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// FactOptions configures a fact extraction run.
type FactOptions struct {
	// Directory is the working directory for loading; empty means the
	// process working directory.
	Directory string
	// Patterns are the package patterns to load (e.g. "./...").
	Patterns []string
	// IncludeTests loads test files and test packages too.
	IncludeTests bool
	// IncludeGenerated keeps declarations from generated files.
	IncludeGenerated bool
	// BuildTags are extra build tags for loading.
	BuildTags []string
	// Workers bounds extraction concurrency (0 = min(GOMAXPROCS, packages)).
	Workers int
	// ContinueOnError skips packages with load or type errors instead of
	// failing the run.
	ContinueOnError bool
}

// FactSource loads a project once and extracts its source facts.
type FactSource interface {
	// Load returns the main module path (empty when unknown) and one
	// PackageExtract per analyzed package, honoring the ordering contract
	// documented on domain.PackageExtract.
	Load(
		ctx context.Context,
		opts FactOptions,
	) (modulePath string, packages []domain.PackageExtract, err error)
}
