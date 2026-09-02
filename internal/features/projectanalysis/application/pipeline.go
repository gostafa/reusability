package application

import (
	"context"

	cohesion "github.com/gostafa/reusability/internal/features/cohesion/application"
	complexity "github.com/gostafa/reusability/internal/features/complexity/application"
	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	reusability "github.com/gostafa/reusability/internal/features/reusability/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/internal/shared/workerpool"
)

// runWorkers is a seam so tests can force workerpool.Run failures.
var runWorkers = workerpool.Run

// Pipeline implements the inbound Analyzer port.
type Pipeline struct {
	facts typefacts.Collector
}

// NewPipeline returns a pipeline backed by the given fact collector.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return &Pipeline{facts: facts}
}

var _ inbound.Analyzer = (*Pipeline)(nil)

// Analyze runs the full pipeline for one request.
func (p *Pipeline) Analyze(ctx context.Context, opts inbound.Options) (inbound.Result, error) {
	// Reusability is the only reported metric; its inputs are computed
	// internally through the dependency closure and never rendered.
	reported := []string{metrics.MetricReusability}
	compute := nameSet(metrics.Closure(reported))

	// Reusability weights are validated up front so a bad configuration
	// fails before any loading happens.
	reusabilityCalculator, err := newReusabilityCalculator(compute, opts.Weights)
	if err != nil {
		return inbound.Result{}, err
	}

	facts, err := p.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return inbound.Result{}, err
	}

	return assembleResult(ctx, &facts, reusabilityCalculator, compute, opts)
}

// assembleResult computes every package's results in parallel, honouring
// cancellation, and folds them into the final report.
func assembleResult(
	ctx context.Context,
	facts *tfdomain.ProjectFacts,
	reusabilityCalculator reusability.Calculator,
	compute map[string]bool,
	opts inbound.Options,
) (inbound.Result, error) {
	if err := ctx.Err(); err != nil {
		return inbound.Result{}, err
	}

	packageResults := make([]inbound.PackageResult, len(facts.Packages))
	workers := workerpool.Workers(opts.Workers, len(facts.Packages))

	err := runWorkers(ctx, workers, len(facts.Packages), func(i int) error {
		packageResults[i] = analyzePackage(
			facts,
			i,
			reusabilityCalculator,
			compute,
			opts,
		)

		return nil
	})
	if err != nil {
		return inbound.Result{}, err
	}

	return inbound.Result{ModulePath: facts.ModulePath, Packages: packageResults}, nil
}

// newReusabilityCalculator builds the reusability calculator when the compute
// set needs it; a nil calculator disables per-type reusability and CBO.
func newReusabilityCalculator(
	compute map[string]bool,
	weights metrics.ReusabilityWeights,
) (reusability.Calculator, error) {
	if !compute[metrics.MetricReusability] && !compute[metrics.MetricCBO] {
		return nil, nil
	}

	return reusability.NewService(weights)
}

// collectOptions maps the analysis request onto the fact-source options.
func collectOptions(opts inbound.Options) tfoutbound.FactOptions {
	return tfoutbound.FactOptions{
		Directory:        opts.Directory,
		Patterns:         opts.Patterns,
		IncludeTests:     opts.IncludeTests,
		IncludeGenerated: opts.IncludeGenerated,
		BuildTags:        opts.BuildTags,
		Workers:          opts.Workers,
		ContinueOnError:  opts.ContinueOnError,
	}
}

// analyzePackage computes one package's type reusability results. It writes
// only into its own result value, so package workers never share mutable
// state.
func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	reusabilityCalculator reusability.Calculator,
	compute map[string]bool,
	opts inbound.Options,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]

	result := inbound.PackageResult{Path: pkg.Path}

	needComplexity := compute[metrics.MetricAMC]
	needCohesion := compute[metrics.MetricLCOM] || compute[metrics.MetricTCC]

	result.Types = make([]inbound.TypeResult, 0, len(pkg.TypeIDs))
	for _, typeID := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(
			&facts.Types[typeID],
			reusabilityCalculator,
			needComplexity,
			needCohesion,
			opts,
		))
	}

	return result
}

// analyzeType computes one type's reusability metric.
func analyzeType(
	t *tfdomain.TypeFacts,
	reusabilityCalculator reusability.Calculator,
	needComplexity, needCohesion bool,
	opts inbound.Options,
) inbound.TypeResult {
	var complexityResult complexity.Result
	if needComplexity {
		complexityResult = complexity.ComputeForType(t)
	}

	var cohesionResult cohesion.Result
	if needCohesion {
		cohesionResult = cohesion.ComputeForType(t, opts.FieldUsageTransitive)
	}

	var reusabilityResult reusability.Result
	if reusabilityCalculator != nil {
		reusabilityResult = reusabilityCalculator.ComputeForType(
			t, complexityResult.AMC, cohesionResult.LCOM,
		)
	}

	return inbound.TypeResult{
		Name:          t.Name,
		Reusability:   reusabilityResult.Reusability,
	}
}

func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	return set
}
