package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"sync"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/reusability"
	"golang.org/x/tools/go/analysis"
)

// Name is the analyzer and golangci-lint custom linter name.
const Name = "reusability"

// Doc is the short documentation shown by go/analysis tooling.
const Doc = `enforce Go type-reusability policy thresholds

Reports policy violations for the type-level reusability index and
structural budgets. Policy thresholds are configured inline in the
golangci-lint settings block.`

// New returns a go/analysis Analyzer that loads the module once, evaluates the
// reusability policy, and emits diagnostics for the package under analysis.
func New(settings Settings) (*analysis.Analyzer, error) {
	s := settings.withDefaults()
	if err := s.validate(); err != nil {
		return nil, err
	}

	r := newRunner(s)

	return &analysis.Analyzer{
		Name: Name,
		Doc:  Doc,
		Run:  r.run,
	}, nil
}

// reportAnalyzer is the facade behavior needed by the go/analysis runner.
// Keeping it narrow allows the runner to be exercised without loading a real
// module.
type reportAnalyzer interface {
	// Analyze evaluates one reusability configuration.
	Analyze(context.Context, reusability.Config) (reusability.Report, error)
}

// analyzeFunc adapts a function to reportAnalyzer.
type analyzeFunc func(context.Context, reusability.Config) (reusability.Report, error)

// Analyze delegates to the adapted function.
func (f analyzeFunc) Analyze(
	ctx context.Context,
	cfg reusability.Config,
) (reusability.Report, error) {
	return f(ctx, cfg)
}

// runner caches a whole-module analysis so each package pass only filters and
// reports diagnostics. Coupling metrics require a multi-package load.
type runner struct {
	settings Settings
	analyzer reportAnalyzer

	once  sync.Once
	byPkg map[string][]policydomain.Violation
	err   error
}

func newRunner(settings Settings) *runner {
	return &runner{settings: settings, analyzer: analyzeFunc(reusability.Analyze)}
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	r.once.Do(r.load)

	if r.err != nil {
		return nil, r.err
	}

	reportViolations(pass, r.byPkg[pass.Pkg.Path()])

	return nil, nil
}

// reportViolations emits one diagnostic per violation for this package pass.
func reportViolations(pass *analysis.Pass, violations []policydomain.Violation) {
	for _, v := range violations {
		pass.Report(analysis.Diagnostic{
			Pos:      violationPos(pass, v),
			Category: Name,
			Message:  formatViolation(v),
		})
	}
}

func (r *runner) load() {
	r.byPkg, r.err = computeViolations(r.settings, r.analyzer)
}

// computeViolations performs the fallible analysis work before the runner
// caches its result for subsequent package passes.
func computeViolations(
	settings Settings,
	analyzer reportAnalyzer,
) (map[string][]policydomain.Violation, error) {
	policy, err := settings.policy()
	if err != nil {
		return nil, fmt.Errorf("reusability policy: %w", err)
	}

	report, err := analyzer.Analyze(
		context.Background(),
		settings.toConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("reusability analyze: %w", err)
	}

	return groupByPackage(policydomain.Evaluate(report, policy)), nil
}

func groupByPackage(violations []policydomain.Violation) map[string][]policydomain.Violation {
	byPkg := make(map[string][]policydomain.Violation, len(violations))
	for _, v := range violations {
		byPkg[v.Package] = append(byPkg[v.Package], v)
	}

	return byPkg
}

func violationPos(pass *analysis.Pass, v policydomain.Violation) token.Pos {
	if v.Function != "" {
		return exactFuncPos(pass, v.Type, v.Function)
	}

	if v.Type != "" {
		return typePos(pass, v.Type)
	}

	return structuralPos(pass, v.Key)
}
