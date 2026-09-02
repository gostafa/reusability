// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	reusability "github.com/gostafa/reusability/internal/features/reusability/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/internal/shared/workerpool"
)

// NewPipeline returns a pipeline backed by the given fact collector.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return newPipeline(facts, workerpool.Run)
}

func newPipeline(
	facts typefacts.Collector,
	runWorkers func(context.Context, workerpool.RunConfig) error,
) *Pipeline {
	return &Pipeline{
		analyze: func(ctx context.Context, opts *inbound.Options) (inbound.Result, error) {
			return runPipeline(ctx, &pipelineInput{
				facts: facts, opts: opts, runWorkers: runWorkers,
			})
		},
	}
}

// Analyze runs the full pipeline for one request.
func (pipeline *Pipeline) Analyze(
	ctx context.Context,
	opts *inbound.Options,
) (inbound.Result, error) {
	result, err := pipeline.analyze(ctx, opts)
	if err != nil {
		return inbound.Result{}, fmt.Errorf("pipeline analyze: %w", err)
	}

	return result, nil
}

func runPipeline(ctx context.Context, input *pipelineInput) (inbound.Result, error) {
	compute := nameSet(metrics.Closure([]string{metrics.MetricReusability}))

	calculator, err := newReusabilityCalculator(compute, &input.opts.Weights)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	return runPreparedPipeline(ctx, input, calculator)
}

func runPreparedPipeline(
	ctx context.Context,
	input *pipelineInput,
	calculator *reusability.Service,
) (inbound.Result, error) {
	projectFacts, err := loadFacts(ctx, input.facts, input.opts)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	result, err := assembleResult(ctx, input, projectFacts, calculator)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	return result, nil
}

func loadFacts(
	ctx context.Context,
	collector typefacts.Collector,
	opts *inbound.Options,
) (*tfdomain.ProjectFacts, error) {
	fo := collectOptions(opts)

	facts, err := collector.Collect(ctx, &fo)
	if err != nil {
		return nil, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	return &facts, nil
}

func analyzePackage(input *packageAnalysisInput) inbound.PackageResult {
	pkg := &input.facts.Packages[input.pkgID]
	result := inbound.PackageResult{Path: pkg.Path}

	result.Types = make([]inbound.TypeResult, zero, len(pkg.TypeIDs))

	for index := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(
			&input.facts.Types[pkg.TypeIDs[index]], input.calculator, input.pipeline.opts,
		))
	}

	return result
}

func analyzeType(
	typeFacts *tfdomain.TypeFacts,
	calculator *reusability.Service,
	opts *inbound.Options,
) inbound.TypeResult {
	return inbound.TypeResult{
		Name:        typeFacts.Name,
		Reusability: typeReusability(calculator, typeFacts, fieldUsageMode(opts)),
	}
}

func assembleResult(
	ctx context.Context,
	input *pipelineInput,
	facts *tfdomain.ProjectFacts,
	calculator *reusability.Service,
) (inbound.Result, error) {
	err := ctx.Err()
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	packageResults, err := fillPackageResults(ctx, input, facts, calculator)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	return inbound.Result{ModulePath: facts.ModulePath, Packages: packageResults}, nil
}

func collectOptions(opts *inbound.Options) tfoutbound.FactOptions {
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

func fieldUsageMode(opts *inbound.Options) string {
	if opts.FieldUsageTransitive {
		return fieldUsageTransitive
	}

	return fieldUsageDirect
}

func fillPackageResults(
	ctx context.Context,
	input *pipelineInput,
	facts *tfdomain.ProjectFacts,
	calculator *reusability.Service,
) ([]inbound.PackageResult, error) {
	packageResults := make([]inbound.PackageResult, zero, len(facts.Packages))

	for range facts.Packages {
		packageResults = append(packageResults, inbound.PackageResult{})
	}

	err := input.runWorkers(ctx, packageWorkerConfig(&workerConfigInput{
		pipeline: input, facts: facts, calculator: calculator, packageResults: packageResults,
	}))
	if err != nil {
		return nil, fmt.Errorf("fill packages: %w", err)
	}

	return packageResults, nil
}

func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))

	for index := range names {
		set[names[index]] = true
	}

	return set
}

func newReusabilityCalculator(
	compute map[string]bool,
	weights *metrics.ReusabilityWeights,
) (*reusability.Service, error) {
	resolved := weights

	if !compute[metrics.MetricReusability] && !compute[metrics.MetricCBO] {
		resolved = nil
	}

	service, err := reusability.NewService(resolved)
	if err != nil {
		return nil, fmt.Errorf(errFmtOp, opNewCalc, err)
	}

	return service, nil
}

func packageWorkerConfig(input *workerConfigInput) workerpool.RunConfig {
	return workerpool.RunConfig{
		Workers:   workerpool.Workers(input.pipeline.opts.Workers, len(input.facts.Packages)),
		TaskCount: len(input.facts.Packages),
		Fn: func(index int) error {
			input.packageResults[index] = analyzePackage(&packageAnalysisInput{
				pipeline:   input.pipeline,
				facts:      input.facts,
				calculator: input.calculator,
				pkgID:      index,
			})

			return nil
		},
	}
}

func typeReusability(
	calculator *reusability.Service,
	typeFacts *tfdomain.TypeFacts,
	fieldUsage string,
) metrics.MetricResult {
	if calculator == nil {
		return metrics.MetricResult{}
	}

	return calculator.ComputeForType(typeFacts, fieldUsage).Reusability
}
