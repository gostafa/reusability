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
	return &Pipeline{facts: facts}
}

// Analyze runs the full pipeline for one request.
func (pipeline *Pipeline) Analyze(
	ctx context.Context,
	opts *inbound.Options,
) (inbound.Result, error) {
	compute := nameSet(metrics.Closure([]string{metrics.MetricReusability}))

	calculator, err := newReusabilityCalculator(compute, &opts.Weights)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	facts, err := loadFacts(ctx, pipeline.facts, opts)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	result, err := assembleResult(ctx, facts, calculator, opts)
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

func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	calculator *reusability.Service,
	opts *inbound.Options,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]
	result := inbound.PackageResult{Path: pkg.Path}

	result.Types = make([]inbound.TypeResult, zero, len(pkg.TypeIDs))

	for index := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(
			&facts.Types[pkg.TypeIDs[index]], calculator, opts,
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
	facts *tfdomain.ProjectFacts,
	calculator *reusability.Service,
	opts *inbound.Options,
) (inbound.Result, error) {
	err := ctx.Err()
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	packageResults, err := fillPackageResults(ctx, facts, calculator, opts)
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
	facts *tfdomain.ProjectFacts,
	calculator *reusability.Service,
	opts *inbound.Options,
) ([]inbound.PackageResult, error) {
	packageResults := make([]inbound.PackageResult, zero, len(facts.Packages))

	for range facts.Packages {
		packageResults = append(packageResults, inbound.PackageResult{})
	}

	err := runWorkers(ctx, packageWorkerConfig(facts, calculator, opts, packageResults))
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

func packageWorkerConfig(
	facts *tfdomain.ProjectFacts,
	calculator *reusability.Service,
	opts *inbound.Options,
	packageResults []inbound.PackageResult,
) workerpool.RunConfig {
	return workerpool.RunConfig{
		Workers:   workerpool.Workers(opts.Workers, len(facts.Packages)),
		TaskCount: len(facts.Packages),
		Fn: func(index int) error {
			packageResults[index] = analyzePackage(facts, index, calculator, opts)

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
