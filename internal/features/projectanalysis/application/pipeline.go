package application

import (
	"context"

	cohesion "github.com/gostafa/reusability/internal/features/cohesion/application"
	complexity "github.com/gostafa/reusability/internal/features/complexity/application"
	complexitydomain "github.com/gostafa/reusability/internal/features/complexity/domain"
	coupling "github.com/gostafa/reusability/internal/features/coupling/domain"
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
	display := nameSet(reported)
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

	return assembleResult(ctx, &facts, reusabilityCalculator, display, compute, opts)
}

// assembleResult computes every package's results in parallel, honouring
// cancellation, and folds them into the final report.
func assembleResult(
	ctx context.Context,
	facts *tfdomain.ProjectFacts,
	reusabilityCalculator reusability.Calculator,
	display, compute map[string]bool,
	opts inbound.Options,
) (inbound.Result, error) {
	if err := ctx.Err(); err != nil {
		return inbound.Result{}, err
	}

	// The dependency graph is cheap and feeds the structural Ca/Ce facts,
	// so it is built regardless of the selected metrics.
	graph := coupling.BuildDependencyGraph(facts, coupling.Scope(opts.DependencyScope))

	packageResults := make([]inbound.PackageResult, len(facts.Packages))
	workers := workerpool.Workers(opts.Workers, len(facts.Packages))

	err := runWorkers(ctx, workers, len(facts.Packages), func(i int) error {
		packageResults[i] = analyzePackage(
			facts,
			i,
			graph.Coupling(i),
			reusabilityCalculator,
			display,
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

// analyzePackage computes one package's display metrics and those of its
// types. It writes only into its own result value, so package workers never
// share mutable state.
func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	pkgCoupling coupling.Coupling,
	reusabilityCalculator reusability.Calculator,
	display, compute map[string]bool,
	opts inbound.Options,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]

	result := inbound.PackageResult{
		Path:            pkg.Path,
		Afferent:        pkgCoupling.Afferent,
		Efferent:        pkgCoupling.Efferent,
		ExportedFuncs:   pkg.ExportedFuncCount,
		UnexportedFuncs: pkg.UnexportedFuncCount,
		Vars:            pkg.VarCount,
		Consts:          pkg.ConstCount,
		Variables:       declarationResults(pkg.Variables),
		Constants:       declarationResults(pkg.Constants),
		Functions:       functionResults(pkg.Functions),
	}

	needComplexity := compute[metrics.MetricAMC]
	needCohesion := compute[metrics.MetricLCOM] || compute[metrics.MetricTCC]

	result.Types = make([]inbound.TypeResult, 0, len(pkg.TypeIDs))
	for _, typeID := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(
			&facts.Types[typeID],
			reusabilityCalculator,
			display,
			needComplexity,
			needCohesion,
			opts,
		))
	}

	return result
}

// analyzeType computes one type's displayed metrics in the fixed metric
// order.
func analyzeType(
	t *tfdomain.TypeFacts,
	reusabilityCalculator reusability.Calculator,
	display map[string]bool,
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

	typeResult := inbound.TypeResult{
		Name:          t.Name,
		Exported:      t.Exported,
		Kind:          t.Kind,
		Pos:           t.Pos,
		Fields:        len(t.Fields),
		FieldDetails:  append([]tfdomain.FieldFacts(nil), t.Fields...),
		Methods:       len(t.Methods),
		MethodDetails: methodResults(t),
	}

	for _, name := range metrics.TypeMetricOrder() {
		if !display[name] {
			continue
		}

		switch name {
		case metrics.MetricAMC:
			typeResult.Metrics = append(typeResult.Metrics, complexityResult.AMC)
		case metrics.MetricLCOM:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.LCOM)
		case metrics.MetricTCC:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.TCC)
		case metrics.MetricCBO:
			typeResult.Metrics = append(typeResult.Metrics, reusabilityResult.CBO)
		case metrics.MetricReusability:
			typeResult.Metrics = append(typeResult.Metrics, reusabilityResult.Reusability)
		}
	}

	return typeResult
}

func declarationResults(decls []tfdomain.DeclarationFacts) []inbound.DeclarationResult {
	if len(decls) == 0 {
		return nil
	}

	out := make([]inbound.DeclarationResult, len(decls))
	for i, d := range decls {
		out[i] = inbound.DeclarationResult{
			Name:     d.Name,
			Exported: d.Exported,
			Pos:      d.Pos,
		}
	}

	return out
}

func functionResults(functions []tfdomain.FunctionFacts) []inbound.FunctionResult {
	if len(functions) == 0 {
		return nil
	}

	out := make([]inbound.FunctionResult, len(functions))
	for i, fn := range functions {
		out[i] = inbound.FunctionResult{
			Name:       fn.Name,
			Exported:   fn.Exported,
			Pos:        fn.Pos,
			Lines:      fn.Lines,
			Cyclomatic: complexitydomain.Cyclomatic(fn.Branches),
			Branches:   fn.Branches,
		}
	}

	return out
}

func methodResults(t *tfdomain.TypeFacts) []inbound.FunctionResult {
	if len(t.Methods) == 0 {
		return nil
	}

	out := make([]inbound.FunctionResult, len(t.Methods))
	for i, method := range t.Methods {
		out[i] = inbound.FunctionResult{
			Name:       method.Name,
			Exported:   method.Exported,
			Receiver:   t.Name,
			Pos:        method.Pos,
			Lines:      method.Lines,
			Cyclomatic: complexitydomain.Cyclomatic(method.Branches),
			Branches:   method.Branches,
		}
	}

	return out
}

func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	return set
}
