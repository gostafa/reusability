// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gostafa/reusability"
	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

func ptr(value float64) *float64 {
	return &value
}

func TestSettingsDecodingAndValidation(t *testing.T) {
	t.Parallel()
	var settings Settings
	if err := UnmarshalSettings(
		[]byte(
			`{"field-usage":"transitive","reusability-weights":{"cohesion":0.1,"coupling":0.2,"testability":0.3,"documentation":0.4}}`,
		),
		&settings,
	); err != nil {
		t.Fatal(err)
	}
	weights := settingsReusabilityWeights(&settings)
	if weights.Cohesion != 0.1 || weights.Documentation != 0.4 ||
		settings.FieldUsage != "transitive" {
		t.Fatalf("settings = %+v; weights = %+v", settings, weights)
	}
	settings.ReusabilityWeights = &ReusabilityWeightSettings{Coupling: ptr(0.2)}
	if settingsReusabilityWeights(&settings).Coupling != 0.2 {
		t.Fatal("partial overrides lost")
	}
	for _, data := range []string{`{`, `{"unknown":true}`, `{"workers":"many"}`} {
		if err := UnmarshalSettings([]byte(data), &settings); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	if _, err := marshalRemappedKeys(
		map[string]json.RawMessage{"bad": []byte(`invalid`)},
	); err == nil {
		t.Fatal("invalid raw JSON accepted")
	}
	for _, invalid := range []Settings{
		{DependencyScope: "invalid"},
		{FieldUsage: "invalid"},
		{ReusabilityWeights: &ReusabilityWeightSettings{Cohesion: ptr(-1)}},
		{Rules: []RuleSettings{{Pattern: "**"}}},
		{Rules: []RuleSettings{{Pattern: "", Min: ptr(0.5)}}},
	} {
		if _, err := New(&invalid); err == nil {
			t.Fatalf("accepted %+v", invalid)
		}
	}
}

func TestAnalysisRunnerAndFallbackPositions(t *testing.T) {
	t.Parallel()
	settings := settingsWithDefaults(&Settings{})
	state := newRunnerWith(
		&settings,
		func(_ context.Context, _ *reusability.Config) (reusability.Report, error) {
			return reusability.Report{
				Packages: []reusability.PackageReport{
					{
						Path: "example/p",
						Types: []reusability.TypeReport{
							{
								Name:        "T",
								Reusability: reusability.MetricResult{Applicable: true, Value: 0.1},
							},
						},
					},
				},
			}, nil
		},
	)
	pass := &analysis.Pass{
		Pkg:    types.NewPackage("example/p", "p"),
		Report: func(analysis.Diagnostic) {},
	}
	if _, err := runRunner(state, pass); err != nil {
		t.Fatal(err)
	}
	byPkg, getErr := state.get(t.Context())
	if getErr != nil || len(byPkg["example/p"]) != 1 {
		t.Fatalf("violations = %+v, err = %v", byPkg, getErr)
	}
	settings.Directory = filepath.Join(t.TempDir(), "missing")
	created, err := New(&settings)
	if err != nil {
		t.Fatal(err)
	}
	// Cache an error to exercise the analyzer's registered Run callback.
	state = newRunnerWith(
		&settings,
		func(context.Context, *reusability.Config) (reusability.Report, error) {
			return reusability.Report{}, errors.New("cached")
		},
	)
	if _, err := runRunner(state, pass); err == nil {
		t.Fatal("cached error was lost")
	}
	if _, err := created.Run(pass); err == nil {
		t.Fatal("missing package analysis should fail")
	}
	pass.Files = []*ast.File{nil, {Package: token.Pos(9)}}
	if typePos(pass, "Missing") != token.Pos(9) {
		t.Fatal("missing type must use package position")
	}
	if formatNumber(1) != "1" {
		t.Fatal("integer formatting changed")
	}
}

func TestNewDeclaresRunResultType(t *testing.T) {
	created, err := New(&Settings{})
	if err != nil {
		t.Fatal(err)
	}

	if created.ResultType == nil || created.ResultType.String() != "*analyzer.runResult" {
		t.Fatalf("ResultType = %v, want *analyzer.runResult", created.ResultType)
	}
}

func TestRunnerRunReportsPackageAndTypeViolations(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	byPkg := map[string][]policydomain.Violation{
		"example.com/p": {{
			Package:   "example.com/p",
			Type:      "Widget",
			Value:     0.5,
			Threshold: 0.8,
			Rule:      "**",
		}},
	}
	r := &runner{get: func(context.Context) (map[string][]policydomain.Violation, error) {
		return byPkg, nil
	}}

	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Fset:   fset,
		Files:  []*ast.File{file},
		Pkg:    types.NewPackage("example.com/p", "p"),
		Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
	}

	if _, err := runRunner(r, pass); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if !strings.Contains(diagnostics[0].Message, "is below min 0.8") {
		t.Errorf("type diagnostic = %q", diagnostics[0].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{get: func(context.Context) (map[string][]policydomain.Violation, error) {
		return nil, sentinel
	}}
	if _, err := runRunner(failing, pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	min := 2.0
	settings := Settings{Rules: []RuleSettings{{Pattern: "**", Min: &min}}}
	s := settingsWithDefaults(&settings)
	r := newRunner(&s)
	_, err := r.get(t.Context())
	if err == nil || !strings.Contains(err.Error(), "reusability policy") {
		t.Fatalf("policy load error = %v", err)
	}

	settings = Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Patterns:  []string{defaultPackagePattern},
	}
	s = settingsWithDefaults(&settings)
	r = newRunner(&s)
	_, err = r.get(t.Context())
	if err == nil || !strings.Contains(err.Error(), "reusability analyze") {
		t.Fatalf("analysis load error = %v", err)
	}
}

func TestRunnerSharesOneAnalysisPerInstance(t *testing.T) {
	t.Parallel()
	var calls, diagnostics atomic.Int32
	analyze := func(ctx context.Context, _ *reusability.Config) (reusability.Report, error) {
		calls.Add(1)
		if ctx == nil {
			t.Error("factory context was lost")
		}
		return reusability.Report{Packages: []reusability.PackageReport{{
			Path: "example/p", Types: []reusability.TypeReport{{
				Name: "Widget", Reusability: reusability.MetricResult{Applicable: true, Value: 0.1},
			}},
		}}}, nil
	}
	settings := settingsWithDefaults(&Settings{})
	state := newRunnerWith(&settings, analyze)
	var workers sync.WaitGroup
	for range 32 {
		workers.Go(func() {
			pass := &analysis.Pass{
				Pkg:    types.NewPackage("example/p", "p"),
				Report: func(analysis.Diagnostic) { diagnostics.Add(1) },
			}
			if _, err := runRunner(state, pass); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	if calls.Load() != 1 || diagnostics.Load() != 32 {
		t.Fatalf("calls = %d, diagnostics = %d", calls.Load(), diagnostics.Load())
	}
	other := newRunnerWith(&settings, analyze)
	if _, err := other.get(t.Context()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("separate instance shared initialization: calls = %d", calls.Load())
	}
}

func TestRunnerSnapshotsSettings(t *testing.T) {
	t.Parallel()
	settings := settingsWithDefaults(&Settings{
		Patterns: []string{"./original"}, BuildTags: []string{"original"},
		Rules: []RuleSettings{{Pattern: "**", Min: ptr(0.8)}},
		ReusabilityWeights: &ReusabilityWeightSettings{
			Cohesion: ptr(0.1), Coupling: ptr(0.2), Testability: ptr(0.3), Documentation: ptr(0.4),
		},
	})
	state := newRunnerWith(
		&settings,
		func(_ context.Context, cfg *reusability.Config) (reusability.Report, error) {
			if cfg.Patterns[0] != "./original" || cfg.BuildTags[0] != "original" ||
				cfg.ReusabilityWeights != (reusability.Weights{Cohesion: 0.1, Coupling: 0.2, Testability: 0.3, Documentation: 0.4}) {
				t.Errorf("settings were mutated: %+v", cfg)
			}
			return reusability.Report{Packages: []reusability.PackageReport{{
				Path: "example/p", Types: []reusability.TypeReport{
					{
						Name:        "Widget",
						Reusability: reusability.MetricResult{Applicable: true, Value: 0.5},
					},
				},
			}}}, nil
		},
	)
	settings.Patterns[0] = "./mutated"
	settings.BuildTags[0] = "mutated"
	settings.Rules[0].Pattern = "elsewhere"
	*settings.Rules[0].Min = 0.1
	*settings.ReusabilityWeights.Cohesion = 1
	*settings.ReusabilityWeights.Coupling = 1
	*settings.ReusabilityWeights.Testability = 1
	*settings.ReusabilityWeights.Documentation = 1
	byPkg, err := state.get(t.Context())
	if err != nil || len(byPkg["example/p"]) != 1 || byPkg["example/p"][0].Threshold != 0.8 {
		t.Fatalf("cached policy was mutated: %+v, %v", byPkg, err)
	}
}

func TestSnapshotOptionalValues(t *testing.T) {
	t.Parallel()
	settings := Settings{
		Rules:              []RuleSettings{{Pattern: "**"}},
		ReusabilityWeights: &ReusabilityWeightSettings{Cohesion: ptr(0.3)},
	}
	snapshot := snapshotSettings(&settings)
	if snapshot.Rules[0].Min != nil || snapshot.ReusabilityWeights.Coupling != nil ||
		snapshot.ReusabilityWeights.Testability != nil || snapshot.ReusabilityWeights.Documentation != nil {
		t.Fatalf("unset overrides changed: %+v", snapshot)
	}
}

func TestComputeViolationsForwardsContext(t *testing.T) {
	t.Parallel()
	settings := settingsWithDefaults(&Settings{})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	analyze := func(got context.Context, _ *reusability.Config) (reusability.Report, error) {
		if got != ctx {
			t.Error("analysis did not receive the factory context")
		}
		return reusability.Report{}, nil
	}
	if _, err := computeViolations(ctx, &settings, analyze); err != nil {
		t.Fatal(err)
	}
}

func TestInlinePolicyDefaultsAndIgnoresModularityFile(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(
		config,
		[]byte("version: 1\npackage:\n  types: 3\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	defaultsSettings := Settings{Directory: dir}
	defaults, err := settingsRules(&defaultsSettings)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults) != 1 || defaults[0].Pattern != "**" || defaults[0].Min != 0.7 {
		t.Fatalf("default rules = %+v", defaults)
	}

	min := 0.6
	inlineSettings := Settings{Rules: []RuleSettings{
		{Pattern: "**/internal/**", Min: ptr(0.8)},
		{Pattern: "**", Min: &min},
	}}
	inline, err := settingsRules(&inlineSettings)
	if err != nil {
		t.Fatal(err)
	}
	if len(inline) != 2 || inline[0].Min != 0.8 || inline[1].Min != 0.6 {
		t.Fatalf("inline rules = %+v", inline)
	}
}

func TestEmptyPackagePosition(t *testing.T) {
	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}

func TestEmptyResolvedRules(t *testing.T) {
	t.Parallel()
	if err := validateRulesOutcome(nil, nil); !errors.Is(err, errRuleMinRequired) {
		t.Fatalf("empty rules = %v", err)
	}
}

func TestRemapMarshalFailure(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("marshal failed")
	if _, err := remapKebabKeysWith(
		[]byte(`{}`),
		func(map[string]json.RawMessage) ([]byte, error) { return nil, sentinel },
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("remap = %v", err)
	}
}
