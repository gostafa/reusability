package reusability

import (
	"context"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
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
			Path:            pkg.Path,
			Afferent:        pkg.Afferent,
			Efferent:        pkg.Efferent,
			ExportedFuncs:   pkg.ExportedFuncs,
			UnexportedFuncs: pkg.UnexportedFuncs,
			Vars:            pkg.Vars,
			Consts:          pkg.Consts,
			Variables:       declarationReports(pkg.Variables),
			Constants:       declarationReports(pkg.Constants),
			Functions:       functionReports(pkg.Functions),
			Types:           make([]TypeReport, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			out.Types[j] = TypeReport{
				Name:          t.Name,
				Exported:      t.Exported,
				Kind:          typeKindName(t.Kind),
				Position:      positionReport(t.Pos),
				Fields:        t.Fields,
				FieldDetails:  fieldReports(t.FieldDetails),
				Methods:       t.Methods,
				MethodDetails: functionReports(t.MethodDetails),
				Metrics:       t.Metrics,
			}
		}

		report.Packages[i] = out
	}

	return report, nil
}

func declarationReports(decls []inbound.DeclarationResult) []DeclarationReport {
	if len(decls) == 0 {
		return nil
	}

	out := make([]DeclarationReport, len(decls))
	for i, d := range decls {
		out[i] = DeclarationReport{
			Name:     d.Name,
			Exported: d.Exported,
			Position: positionReport(d.Pos),
		}
	}

	return out
}

func fieldReports(fields []tfdomain.FieldFacts) []FieldReport {
	if len(fields) == 0 {
		return nil
	}

	out := make([]FieldReport, len(fields))
	for i, f := range fields {
		out[i] = FieldReport{Name: f.Name, Exported: f.Exported, Embedded: f.Embedded}
	}

	return out
}

func functionReports(functions []inbound.FunctionResult) []FunctionReport {
	if len(functions) == 0 {
		return nil
	}

	out := make([]FunctionReport, len(functions))
	for i, fn := range functions {
		out[i] = FunctionReport{
			Name:       fn.Name,
			Exported:   fn.Exported,
			Receiver:   fn.Receiver,
			Position:   positionReport(fn.Pos),
			Lines:      fn.Lines,
			Cyclomatic: fn.Cyclomatic,
			Branches:   branchStatsReport(fn.Branches),
		}
	}

	return out
}

func positionReport(pos tfdomain.Position) Position {
	return Position{File: pos.File, Line: pos.Line, Column: pos.Column}
}

func branchStatsReport(branches tfdomain.BranchStats) BranchStats {
	return BranchStats{
		Ifs:         branches.Ifs,
		Fors:        branches.Fors,
		Ranges:      branches.Ranges,
		Cases:       branches.Cases,
		SelectComms: branches.SelectComms,
		LogicalOps:  branches.LogicalOps,
	}
}

func typeKindName(kind tfdomain.TypeKind) string {
	switch kind {
	case tfdomain.KindStruct:
		return "struct"
	case tfdomain.KindInterface:
		return "interface"
	default:
		return "other"
	}
}
