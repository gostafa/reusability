// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/internal/shared/workerpool"
)

type coverageSource struct{}

func (coverageSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return "example.com/m", []tfdomain.PackageExtract{{
		Path: "example.com/m/a", InModule: true,
		Types: []tfdomain.TypeExtract{
			{
				Name:     "A",
				Exported: true,
				Kind:     tfdomain.KindStruct,
				Methods: []tfdomain.MethodFacts{
					{Name: "Do", Exported: true, Branches: tfdomain.BranchStats{Ifs: 1}},
				},
			},
		},
	}}, nil
}

func TestAssembleResultWorkerError(t *testing.T) {
	sentinel := errors.New("workers failed")
	pipeline := pipelineWithWorkers(
		typefacts.NewService(coverageSource{}),
		func(context.Context, workerpool.RunConfig) error {
			return sentinel
		},
	)

	_, err := pipeline(context.Background(), &inbound.Options{
		Patterns: []string{"./..."},
		Weights:  metrics.DefaultReusabilityWeights(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Analyze error = %v, want sentinel", err)
	}
}

func TestReportedMetricIsReusability(t *testing.T) {
	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	result, err := pipeline(context.Background(), &inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
		Weights:         metrics.DefaultReusabilityWeights(),
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	for _, typ := range pkg.Types {
		if typ.Reusability.Name != metrics.MetricReusability {
			t.Fatalf("unexpected type metric %q in display set", typ.Reusability.Name)
		}
	}
}

// White-box: the request→fact-options mapping.
func TestCollectOptionsMapping(t *testing.T) {
	t.Parallel()

	fo := collectOptions(&inbound.Options{
		Directory: "d", Patterns: []string{"./..."}, IncludeTests: true,
		IncludeGenerated: true, BuildTags: []string{"tag"}, Workers: 3, ContinueOnError: true,
	})
	if fo.Directory != "d" || !fo.IncludeTests || !fo.IncludeGenerated ||
		fo.Workers != 3 || !fo.ContinueOnError || len(fo.BuildTags) != 1 {
		t.Fatalf("collectOptions = %+v", fo)
	}
}

// White-box: the reusability service is built only when needed.
func TestNewReusabilityCalculatorGating(t *testing.T) {
	t.Parallel()

	defaults := metrics.DefaultReusabilityWeights()
	calculator, err := newReusabilityCalculator(
		map[string]bool{},
		&defaults,
	)
	if err != nil || calculator == nil {
		t.Fatalf("no reusability/cbo selected → default service; got %v err %v", calculator, err)
	}

	calculator, err = newReusabilityCalculator(
		map[string]bool{metrics.MetricReusability: true},
		&defaults,
	)
	if err != nil || calculator == nil {
		t.Fatalf("reusability selected → calculator; got %v err %v", calculator, err)
	}

	bad := metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1}
	if _, err := newReusabilityCalculator(map[string]bool{metrics.MetricCBO: true},
		&bad); err == nil {
		t.Fatal("bad weights should fail")
	}
}

// White-box: nameSet builds a membership set.
func TestNameSet(t *testing.T) {
	t.Parallel()

	s := nameSet([]string{"a", "b", "a"})
	if len(s) != 2 || !s["a"] || !s["b"] {
		t.Fatalf("nameSet = %v", s)
	}
}

type errSource struct{ err error }

func (e errSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return "", nil, e.err
}

// Black-box: a fact-source failure propagates out of Analyze.
func TestPipelineLoadError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("load failed")
	pipeline := NewPipeline(typefacts.NewService(errSource{err: sentinel}))

	_, err := pipeline(context.Background(), &inbound.Options{
		Patterns: []string{"./..."},
		Weights:  metrics.DefaultReusabilityWeights(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}
}

// Black-box: invalid reusability weights fail before any loading happens.
func TestPipelineBadWeights(t *testing.T) {
	t.Parallel()

	pipeline := NewPipeline(typefacts.NewService(fakeSource{mod: "example.com/m"}))

	_, err := pipeline(context.Background(), &inbound.Options{
		Patterns: []string{"./..."},
		Weights:  metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1},
	})
	if err == nil {
		t.Fatal("invalid weights should fail the run")
	}
}

// fakeSource feeds canned extracts so the whole pipeline runs without loading
// real packages through go/packages.
type fakeSource struct {
	mod  string
	pkgs []tfdomain.PackageExtract
}

func (f fakeSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return f.mod, f.pkgs, nil
}

// Black-box: the pipeline turns extracts into a deterministic report with the
// selected package- and type-level metrics.
func TestPipelineAnalyzeEndToEnd(t *testing.T) {
	t.Parallel()

	src := fakeSource{
		mod: "example.com/m",
		pkgs: []tfdomain.PackageExtract{
			{
				Path: "example.com/m/a", InModule: true, Imports: []string{"example.com/m/b"},
				Types: []tfdomain.TypeExtract{
					{
						Name:     "A",
						Exported: true,
						Kind:     tfdomain.KindStruct,
						Methods: []tfdomain.MethodFacts{
							{Name: "Do", Exported: true, Branches: tfdomain.BranchStats{Ifs: 1}},
						},
					},
				},
			},
			{
				Path:     "example.com/m/b",
				InModule: true,
				Types: []tfdomain.TypeExtract{
					{Name: "B", Exported: true, Kind: tfdomain.KindInterface},
				},
			},
		},
	}
	pipeline := NewPipeline(typefacts.NewService(src))

	result, err := pipeline(context.Background(), &inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
		Weights:         metrics.DefaultReusabilityWeights(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) != 2 || result.Packages[0].Path != "example.com/m/a" ||
		result.Packages[1].Path != "example.com/m/b" {
		t.Fatalf("packages not sorted by path: %+v", result.Packages)
	}

	pkgA := result.Packages[0]

	if len(pkgA.Types) != 1 || pkgA.Types[0].Name != "A" {
		t.Fatalf("a types = %+v", pkgA.Types)
	}

	if got := pkgA.Types[0].Reusability; !got.Applicable {
		t.Errorf("A reusability = %+v, want applicable", got)
	}

	if len(result.Packages[1].Types) != 1 || result.Packages[1].Types[0].Name != "B" {
		t.Fatalf("b types = %+v", result.Packages[1].Types)
	}
}

// Black-box: a cancelled context aborts before doing work.
func TestPipelineCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pipeline := NewPipeline(typefacts.NewService(fakeSource{mod: "m"}))
	if _, err := pipeline(
		ctx,
		&inbound.Options{Patterns: []string{"./..."}, Weights: metrics.DefaultReusabilityWeights()},
	); err == nil {
		t.Fatal("cancelled context should abort")
	}
}
