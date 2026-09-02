// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func TestRuleListErrorsAndString(t *testing.T) {
	var rules ruleList
	if err := rules.Set(" **/internal/** : 0.7 "); err != nil {
		t.Fatal(err)
	}
	if got := rules.String(); got != "**/internal/**:0.7" {
		t.Fatalf("String() = %q", got)
	}

	for _, value := range []string{"pattern", ":0.7", "pattern=0.7", "pattern:not-a-number"} {
		if err := rules.Set(value); err == nil {
			t.Errorf("Set(%q) succeeded, want error", value)
		}
	}
}

func TestResolvePolicyRequiresRules(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(config, []byte("not: valid\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := resolvePolicy(ruleList{}); err == nil {
		t.Fatal("empty flag policy succeeded")
	}

	rules := ruleList{items: []ruleSpec{{pattern: "**/internal/**", minimum: 0.8}}}
	got, source, err := resolvePolicy(rules)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].minimum != 0.8 || source != policySourceFlagRules {
		t.Fatalf("flag rules = %+v, source = %q", got, source)
	}

	bad := ruleList{items: []ruleSpec{{pattern: "**", minimum: 2}}}
	if _, _, err := resolvePolicy(bad); err == nil {
		t.Fatal("invalid rule min succeeded")
	}
}

func TestRunEarlyErrorPaths(t *testing.T) {
	if code := execute([]string{"--verbose", "--dependency-scope=nope"}); code != 1 {
		t.Fatalf("invalid dependency scope exit = %d, want 1", code)
	}

	if code := execute([]string{"--reusability-weight-cohesion=-1"}); code != 1 {
		t.Fatalf("invalid reusability weight exit = %d, want 1", code)
	}

	if code := execute([]string{
		"--reusability-weight-cohesion=0",
		"--reusability-weight-coupling=0",
		"--reusability-weight-testability=0",
		"--reusability-weight-documentation=0",
	}); code != 1 {
		t.Fatalf("all-zero reusability weights exit = %d, want 1", code)
	}

	badProfile := filepath.Join(t.TempDir(), "missing", "cpu.prof")
	if code := execute([]string{"--cpu-profile=" + badProfile}); code != 1 {
		t.Fatalf("bad CPU profile exit = %d, want 1", code)
	}

	badTemp := filepath.Join(t.TempDir(), "missing")
	t.Setenv("TMPDIR", badTemp)
	t.Setenv("TMP", badTemp)
	if code := runWebHelp(); code != 1 {
		t.Fatalf("web help with bad temp dir exit = %d, want 1", code)
	}
}

func TestRunPassesReusabilityWeights(t *testing.T) {
	var got reusability.Weights
	analyze := func(_ context.Context, cfg *reusability.Config) (reusability.Report, error) {
		got = cfg.ReusabilityWeights

		return reusability.Report{}, nil
	}

	if code := executeWithAnalyze([]string{
		"--reusability-weight-cohesion=0.1",
		"--reusability-weight-coupling=0.2",
		"--reusability-weight-testability=0.3",
		"--reusability-weight-documentation=0.4",
	}, analyze); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}

	want := metrics.ReusabilityWeights{
		Cohesion: 0.1, Coupling: 0.2, Testability: 0.3, Documentation: 0.4,
	}
	if got != want {
		t.Fatalf("weights = %+v, want %+v", got, want)
	}
}

func TestRunCheckRequiresRules(t *testing.T) {
	if code := execute([]string{"--check", "./..."}); code != 2 {
		t.Fatalf("check without rules exit = %d, want 2", code)
	}
}

func TestRunCheckWithRules(t *testing.T) {
	analyze := func(_ context.Context, _ *reusability.Config) (reusability.Report, error) {
		return reusability.Report{Packages: []reusability.PackageReport{{
			Path: "example.com/p",
			Types: []reusability.TypeReport{{
				Name: "T",
				Reusability: metrics.MetricResult{
					Name: metrics.MetricReusability, Applicable: true, Value: 0.5,
				},
			}},
		}}}, nil
	}

	if code := executeWithAnalyze([]string{"--check", `--rule=**:0.8`}, analyze); code != 3 {
		t.Fatalf("policy violation exit = %d, want 3", code)
	}
}

func TestRunCheckSpecificRuleOverridesBaseline(t *testing.T) {
	passingAnalyze := func(_ context.Context, _ *reusability.Config) (reusability.Report, error) {
		return reusability.Report{Packages: []reusability.PackageReport{{
			Path: "example.com/internal/dto",
			Types: []reusability.TypeReport{{Name: "T", Reusability: metrics.MetricResult{
				Name: metrics.MetricReusability, Applicable: true, Value: 0.5,
			}}},
		}}}, nil
	}

	rules := []string{"--check", "--rule=**:0.6", "--rule=**/internal/dto:0.5"}
	if code := executeWithAnalyze(rules, passingAnalyze); code != 0 {
		t.Fatalf("specific override exit = %d, want 0", code)
	}

	failingAnalyze := func(_ context.Context, _ *reusability.Config) (reusability.Report, error) {
		return reusability.Report{Packages: []reusability.PackageReport{{
			Path: "example.com/internal/dto",
			Types: []reusability.TypeReport{{Name: "T", Reusability: metrics.MetricResult{
				Name: metrics.MetricReusability, Applicable: true, Value: 0.49,
			}}},
		}}}, nil
	}
	if code := executeWithAnalyze(rules, failingAnalyze); code != 3 {
		t.Fatalf("regression below override exit = %d, want 3", code)
	}
}

func TestRunPolicySourceLogged(t *testing.T) {
	analyze := func(_ context.Context, _ *reusability.Config) (reusability.Report, error) {
		return reusability.Report{Packages: []reusability.PackageReport{{
			Path: "p",
			Types: []reusability.TypeReport{{
				Name: "T",
				Reusability: metrics.MetricResult{
					Name: metrics.MetricReusability, Applicable: true, Value: 0.9,
				},
			}},
		}}}, nil
	}

	if code := executeWithAnalyze([]string{"--check", `--rule=**:0.7`}, analyze); code != 0 {
		t.Fatalf("passing check exit = %d, want 0", code)
	}
}

func TestResolvePolicyValidatesRules(t *testing.T) {
	specs, _, err := resolvePolicy(ruleList{items: []ruleSpec{
		{pattern: "**/internal/**", minimum: 0.8},
		{pattern: "**", minimum: 0.6},
	}})
	if err != nil {
		t.Fatal(err)
	}

	rules := make([]policydomain.Rule, 0, len(specs))
	for i := range specs {
		rules = append(rules, policydomain.Rule{Pattern: specs[i].pattern, Min: specs[i].minimum})
	}

	if err := policydomain.Validate(rules); err != nil {
		t.Fatal(err)
	}
}

// Black-box: the CLI analyzes the fixture and writes a valid JSON report to
// --output. (Not parallel — it changes the working directory.)
func TestRunFixtureJSON(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixture"))
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	out := filepath.Join(tmp, "report.json")
	cpuProfile := filepath.Join(tmp, "cpu.prof")
	memoryProfile := filepath.Join(tmp, "memory.prof")
	t.Chdir(fixture)

	if code := Run([]string{
		"--format=json",
		"--output=" + out,
		"--cpu-profile=" + cpuProfile,
		"--memory-profile=" + memoryProfile,
		"./...",
	}); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	for _, profile := range []string{cpuProfile, memoryProfile} {
		info, err := os.Stat(profile)
		if err != nil || info.Size() == 0 {
			t.Fatalf("profile %q was not written: info=%v err=%v", profile, info, err)
		}
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON report: %v", err)
	}

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkgs := got["packages"].([]any)
	if len(pkgs) < 7 {
		t.Errorf("packages = %d, want >= 7", len(pkgs))
	}

	first := pkgs[0].(map[string]any)
	for _, key := range []string{"afferent", "efferent", "funcs", "vars", "consts", "functions", "variables", "constants"} {
		if _, ok := first[key]; ok {
			t.Errorf("package should not include removed field %q", key)
		}
	}

	if types := first["types"].([]any); len(types) > 0 {
		typ := types[0].(map[string]any)
		for _, key := range []string{"fields", "methods", "kind", "position", "field_details", "method_details", "metrics", "exported"} {
			if _, ok := typ[key]; ok {
				t.Errorf("type should not include removed field %q", key)
			}
		}
		if _, ok := typ["reusability"]; !ok {
			t.Error("type is missing reusability field")
		}
	}
}

// Black-box: --web writes a self-contained HTML report to --output. (Not
// parallel — it changes the working directory.)
func TestRunFixtureWeb(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixture"))
	if err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "report.html")
	t.Chdir(fixture)

	if code := Run([]string{"--web", "--output=" + out, "./..."}); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	html := string(data)
	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("report does not start with a doctype: %.40q", html)
	}

	if !strings.Contains(html, "example.com/fixture") {
		t.Error("report does not mention the fixture module")
	}
}

// Black-box: --web conflicting with an explicit non-web --format is a usage
// error.
func TestRunWebFormatConflict(t *testing.T) {
	t.Parallel()

	if code := Run([]string{"--web", "--format=json"}); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

// Black-box: --version succeeds.
func TestRunVersion(t *testing.T) {
	t.Parallel()

	if code := Run([]string{"--version"}); code != 0 {
		t.Fatalf("--version exit = %d, want 0", code)
	}
}

// Black-box: --help --web (either order) writes the self-contained metrics
// guide to the OS temp dir and succeeds. The browser never opens here — a
// test process's stdout is a pipe, not a terminal. (Not parallel — it
// changes the temp dir env.)
func TestRunHelpWeb(t *testing.T) {
	for _, args := range [][]string{
		{"--help", "--web"},
		{"--web", "--help"},
		{"-h", "--web"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			t.Setenv("TMP", tmp)

			if code := Run(args); code != 0 {
				t.Fatalf("exit code = %d, want 0", code)
			}

			matches, err := filepath.Glob(filepath.Join(tmp, "reusability-help-*.html"))
			if err != nil {
				t.Fatal(err)
			}

			if len(matches) != 1 {
				t.Fatalf("guide files written = %d, want 1", len(matches))
			}

			data, err := os.ReadFile(matches[0])
			if err != nil {
				t.Fatal(err)
			}

			html := string(data)
			if !strings.HasPrefix(html, "<!doctype html>") {
				t.Errorf("guide does not start with a doctype: %.40q", html)
			}

			for _, want := range []string{`id="docs-data"`, `<math`} {
				if !strings.Contains(html, want) {
					t.Errorf("guide is missing %q", want)
				}
			}
		})
	}
}

// Black-box: plain --help keeps its usage-error exit code.
func TestRunHelpWithoutWeb(t *testing.T) {
	t.Parallel()

	if code := Run([]string{"--help"}); code != 2 {
		t.Fatalf("--help exit = %d, want 2", code)
	}
}

// chdirFixture switches into the sample module used by the policy-gate tests.
// (Not parallel — it changes the working directory.)
func chdirFixture(t *testing.T) {
	t.Helper()

	fixture, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixture"))
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(fixture)
}

// Black-box: a violated rule exits 3.
func TestRunCheckFailsExitsThree(t *testing.T) {
	chdirFixture(t)

	if code := Run(
		[]string{
			"--check",
			`--rule=**:0.99`,
			"--output",
			filepath.Join(t.TempDir(), "r.txt"),
			"./...",
		},
	); code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
}

// Black-box: a satisfiable flag-only policy passes with exit 0.
func TestRunCheckFlagPolicyPasses(t *testing.T) {
	chdirFixture(t)

	args := []string{
		"--check",
		"--output", filepath.Join(t.TempDir(), "r.txt"),
		`--rule=**:0`,
		"./...",
	}

	if code := Run(args); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

// Black-box: an invalid rule is a usage error (exit 2), and gating
// never runs without at least one -rule.
func TestRunCheckKeyAndTriggers(t *testing.T) {
	chdirFixture(t)

	out := filepath.Join(t.TempDir(), "r.txt")

	if code := Run(
		[]string{"--check", `--rule=not-a-number`, "--output", out, "./..."},
	); code != 2 {
		t.Fatalf("invalid rule exit = %d, want 2", code)
	}

	if code := Run([]string{"--check", "--output", out, "./..."}); code != 2 {
		t.Fatalf("empty check exit = %d, want 2", code)
	}

	if code := Run([]string{"--output", out, "./..."}); code != 0 {
		t.Fatalf("ungated exit = %d, want 0", code)
	}
}
