package reusability

import (
	"context"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/reusability/internal/infrastructure/analyzer"
	"github.com/gostafa/reusability/internal/shared/version"
)

// Analyze validates the configuration, runs the analysis pipeline once over
// the configured patterns, and returns a deterministic report. The context
// cancels package loading and metric computation.
func Analyze(ctx context.Context, config Config) (Report, error) {
	cfg := configWithDefaults(config)
	if err := validateConfig(cfg); err != nil {
		return Report{}, err
	}

	result, err := analyzer.NewAnalyzer().Analyze(ctx, inbound.Options{
		Directory:            cfg.Directory,
		Patterns:             cfg.Patterns,
		IncludeTests:         cfg.IncludeTests,
		IncludeGenerated:     cfg.IncludeGenerated,
		BuildTags:            cfg.BuildTags,
		Workers:              cfg.Workers,
		DependencyScope:      string(cfg.DependencyScope),
		FieldUsageTransitive: cfg.FieldUsageMode == FieldUsageTransitive,
		ContinueOnError:      cfg.ContinueOnError,
		Weights:              cfg.ReusabilityWeights,
	})
	if err != nil {
		return Report{}, err
	}

	report := Report{
		SchemaVersion: SchemaVersion,
		Tool:          ToolInfo{Name: ToolName, Version: version.Version},
		Module:        result.ModulePath,
		Packages:      make([]PackageReport, len(result.Packages)),
	}
	for i, pkg := range result.Packages {
		out := PackageReport{
			Path:  pkg.Path,
			Types: make([]TypeReport, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			out.Types[j] = TypeReport{
				Name:          t.Name,
				Reusability:   t.Reusability,
			}
		}

		report.Packages[i] = out
	}

	return report, nil
}
