// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"

	cohesion "github.com/gostafa/reusability/internal/features/cohesion/application"
	cohesdomain "github.com/gostafa/reusability/internal/features/cohesion/domain"
	complexity "github.com/gostafa/reusability/internal/features/complexity/application"
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
	return &Pipeline{facts: facts, runWorkers: workerpool.Run}
}

// Analyze runs the full pipeline for one request.
func (pipeline *Pipeline) Analyze(
	ctx context.Context,
	opts *inbound.Options,
) (inbound.Result, error) {
	// Build compute set, calculator, then collect facts and assemble results.
	compute := nameSet(metrics.Closure([]string{metrics.MetricReusability}))

	calculator, err := newReusabilityCalculator(compute, &opts.Weights)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	result, err := pipeline.collectAndAssemble(ctx, &analysisRun{
		opts: opts, compute: compute, calculator: calculator,
	})
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	return result, nil
}

func (pipeline *Pipeline) collectAndAssemble(
	ctx context.Context,
	run *analysisRun,
) (inbound.Result, error) {
	// Collect facts then assemble package/type metric results.
	facts, err := pipeline.loadFacts(ctx, run.opts)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	result, err := pipeline.assembleFromFacts(ctx, run, facts)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	return result, nil
}

func (pipeline *Pipeline) assembleFromFacts(
	ctx context.Context,
	run *analysisRun,
	facts *tfdomain.ProjectFacts,
) (inbound.Result, error) {
	// Assemble inbound results from collected project facts.
	result, err := assembleResult(ctx, &assembleJob{
		facts:                 facts,
		reusabilityCalculator: run.calculator,
		compute:               run.compute,
		opts:                  run.opts,
		runWorkers:            pipeline.runWorkers,
	})
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	return result, nil
}

func (pipeline *Pipeline) loadFacts(
	ctx context.Context,
	opts *inbound.Options,
) (*tfdomain.ProjectFacts, error) {
	// Map inbound options onto fact-source options and collect.
	fo := collectOptions(opts)

	facts, err := pipeline.facts.Collect(ctx, &fo)
	if err != nil {
		return nil, fmt.Errorf(errFmtOp, opAnalyze, err)
	}

	return &facts, nil
}

func analyzePackage(job *packageJob) inbound.PackageResult {
	pkg := &job.facts.Packages[job.pkgID]
	result := inbound.PackageResult{Path: pkg.Path}
	needs := packageNeeds(job.compute)

	result.Types = make([]inbound.TypeResult, zero, len(pkg.TypeIDs))

	for index := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(&typeJob{
			typeFacts:             &job.facts.Types[pkg.TypeIDs[index]],
			reusabilityCalculator: job.reusabilityCalculator,
			needs:                 needs,
			opts:                  job.opts,
		}))
	}

	return result
}

func analyzeType(job *typeJob) inbound.TypeResult {
	var complexityResult complexity.Result

	if job.needs&needComplexity != zero {
		complexityResult = complexity.ComputeForType(job.typeFacts)
	}

	var cohesionResult cohesion.Result

	if job.needs&needCohesion != zero {
		cohesionResult = cohesion.ComputeForType(job.typeFacts, fieldUsageMode(job.opts))
	}

	return inbound.TypeResult{
		Name:        job.typeFacts.Name,
		Reusability: typeReusability(job, &complexityResult.AMC, &cohesionResult.LCOM),
	}
}

func assembleResult(ctx context.Context, job *assembleJob) (inbound.Result, error) {
	err := ctx.Err()
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	packageResults, err := fillPackageResults(ctx, job)
	if err != nil {
		return inbound.Result{}, fmt.Errorf(errFmtOp, opAssemble, err)
	}

	return inbound.Result{ModulePath: job.facts.ModulePath, Packages: packageResults}, nil
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

func fieldUsageMode(opts *inbound.Options) cohesdomain.FieldUsageMode {
	if opts.FieldUsageTransitive {
		return cohesdomain.FieldUsageTransitive
	}

	return cohesdomain.FieldUsageDirect
}

func fillPackageResults(ctx context.Context, job *assembleJob) ([]inbound.PackageResult, error) {
	packageResults := make([]inbound.PackageResult, zero, len(job.facts.Packages))

	for range job.facts.Packages {
		packageResults = append(packageResults, inbound.PackageResult{})
	}

	err := job.runWorkers(ctx, packageWorkerConfig(job, packageResults))
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
	// Use nil weights (defaults) when reusability metrics are not selected.
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

func packageNeeds(compute map[string]bool) metricNeeds {
	var needs metricNeeds

	if compute[metrics.MetricAMC] {
		needs |= needComplexity
	}

	if compute[metrics.MetricLCOM] || compute[metrics.MetricTCC] {
		needs |= needCohesion
	}

	return needs
}

func packageWorkerConfig(
	job *assembleJob,
	packageResults []inbound.PackageResult,
) workerpool.RunConfig {
	// Fan out one worker task per package into packageResults slots.
	return workerpool.RunConfig{
		Workers:   workerpool.Workers(job.opts.Workers, len(job.facts.Packages)),
		TaskCount: len(job.facts.Packages),
		Fn: func(index int) error {
			packageResults[index] = analyzePackage(&packageJob{
				facts: job.facts, pkgID: index,
				reusabilityCalculator: job.reusabilityCalculator,
				compute:               job.compute,
				opts:                  job.opts,
			})

			return nil
		},
	}
}

func typeReusability(job *typeJob, amc, lcom *metrics.MetricResult) metrics.MetricResult {
	if job.reusabilityCalculator == nil {
		return metrics.MetricResult{}
	}

	return job.reusabilityCalculator.ComputeForType(job.typeFacts, amc, lcom).Reusability
}
