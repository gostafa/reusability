package goloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/workerpool"
	"golang.org/x/tools/go/packages"
)

// loadMode is the exact package information the analysis needs; the project
// is loaded once with this mode.
const loadMode = packages.NeedName |
	packages.NeedModule |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes

// Seams for tests that need to force OS / packages / worker failures.
var (
	packagesLoad      = packages.Load
	runExtractWorkers = workerpool.Run
	osGetwd           = os.Getwd
	filepathAbs       = filepath.Abs
)

// Loader implements outbound.FactSource on top of golang.org/x/tools.
type Loader struct{}

// New returns a compiler-backed fact source.
func New() *Loader { return &Loader{} }

var _ outbound.FactSource = (*Loader)(nil)

// Load loads the requested patterns once, deduplicates test variants, and
// extracts per-package facts with bounded package-level concurrency.
func (l *Loader) Load(
	ctx context.Context,
	opts outbound.FactOptions,
) (string, []domain.PackageExtract, error) {
	return load(ctx, opts)
}

// load implements the fact-source contract; the adapter shell above stays a
// thin delegation.
func load(ctx context.Context, opts outbound.FactOptions) (string, []domain.PackageExtract, error) {
	pkgs, err := loadPackages(ctx, opts)
	if err != nil {
		return "", nil, err
	}

	modulePath := mainModulePath(pkgs)

	extracts, err := extractAll(ctx, pkgs, opts, modulePath)
	if err != nil {
		return "", nil, err
	}

	return modulePath, extracts, nil
}

// loadPackages loads the requested patterns once and selects the loadable,
// deduplicated package set.
func loadPackages(ctx context.Context, opts outbound.FactOptions) ([]*packages.Package, error) {
	patterns := opts.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	cfg := &packages.Config{
		Context: ctx,
		Dir:     opts.Directory,
		Mode:    loadMode,
		Tests:   opts.IncludeTests,
	}
	if len(opts.BuildTags) > 0 {
		cfg.BuildFlags = []string{"-tags=" + strings.Join(opts.BuildTags, ",")}
	}

	loaded, err := packagesLoad(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	if len(loaded) == 0 {
		return nil, fmt.Errorf("no packages matched patterns %v", patterns)
	}

	pkgs, err := selectPackages(loaded, opts.ContinueOnError)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no loadable packages matched patterns %v", patterns)
	}

	return pkgs, nil
}

// extractAll extracts per-package facts with bounded concurrency, releasing
// each package's compiler data as soon as its facts exist.
func extractAll(
	ctx context.Context,
	pkgs []*packages.Package,
	opts outbound.FactOptions,
	modulePath string,
) ([]domain.PackageExtract, error) {
	analyzed := make(map[string]bool, len(pkgs))
	for _, p := range pkgs {
		analyzed[p.PkgPath] = true
	}

	baseDir := resolveBaseDir(opts.Directory)

	extracts := make([]domain.PackageExtract, len(pkgs))
	workers := workerpool.Workers(opts.Workers, len(pkgs))

	err := runExtractWorkers(ctx, workers, len(pkgs), func(i int) error {
		pkg := pkgs[i]
		extracts[i] = extractPackage(pkg, extractorOptions{
			includeGenerated: opts.IncludeGenerated,
			analyzed:         analyzed,
			modulePath:       modulePath,
			baseDir:          baseDir,
		})
		// Release compiler data as soon as the package's facts exist; only
		// this worker touches these fields.
		pkg.Syntax = nil
		pkg.TypesInfo = nil
		pkg.Types = nil

		return nil
	})
	if err != nil {
		return nil, err
	}

	return extracts, nil
}

// selectPackages drops synthesized test binaries, deduplicates test variants
// (preferring the test-augmented variant with more compiled files), and
// applies the error policy.
func selectPackages(loaded []*packages.Package, continueOnError bool) ([]*packages.Package, error) {
	byPath := make(map[string]*packages.Package, len(loaded))

	order := make([]string, 0, len(loaded))
	for _, p := range loaded {
		if strings.HasSuffix(p.PkgPath, ".test") {
			continue // synthesized test main package
		}

		existing, ok := byPath[p.PkgPath]
		if !ok {
			byPath[p.PkgPath] = p
			order = append(order, p.PkgPath)

			continue
		}

		if len(p.CompiledGoFiles) > len(existing.CompiledGoFiles) {
			byPath[p.PkgPath] = p
		}
	}

	var errs []string

	pkgs := make([]*packages.Package, 0, len(order))
	for _, path := range order {
		p := byPath[path]

		broken := len(p.Errors) > 0 || p.Types == nil || p.TypesInfo == nil
		if !broken {
			pkgs = append(pkgs, p)

			continue
		}

		if continueOnError {
			continue
		}

		for _, e := range p.Errors {
			errs = append(errs, fmt.Sprintf("%s: %s", path, e.Msg))
		}

		if len(p.Errors) == 0 {
			errs = append(errs, path+": type information unavailable")
		}
	}

	if len(errs) > 0 {
		const maxShown = 10

		shown := errs
		suffix := ""

		if len(shown) > maxShown {
			shown = shown[:maxShown]
			suffix = fmt.Sprintf("\n… and %d more", len(errs)-maxShown)
		}

		return nil, fmt.Errorf("package load errors (use ContinueOnError to skip):\n%s%s",
			strings.Join(shown, "\n"), suffix)
	}

	return pkgs, nil
}

// mainModulePath returns the path of the main module, when known.
func mainModulePath(pkgs []*packages.Package) string {
	for _, p := range pkgs {
		if p.Module != nil && p.Module.Main {
			return p.Module.Path
		}
	}

	for _, p := range pkgs {
		if p.Module != nil {
			return p.Module.Path
		}
	}

	return ""
}

// resolveBaseDir absolutizes the analysis directory so source positions can
// be reported relative to it.
func resolveBaseDir(dir string) string {
	if dir == "" {
		if wd, err := osGetwd(); err == nil {
			return wd
		}

		return ""
	}

	if abs, err := filepathAbs(dir); err == nil {
		return abs
	}

	return dir
}
