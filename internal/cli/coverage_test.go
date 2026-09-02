package cli

import (
	"context"
	"os"
	"path/filepath"
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

	rules := ruleList{items: []ruleSpec{{pattern: "**/internal/**", min: 0.8}}}
	got, source, err := resolvePolicy(rules)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Min != 0.8 || source != "flag rules" {
		t.Fatalf("flag rules = %+v, source = %q", got, source)
	}

	bad := ruleList{items: []ruleSpec{{pattern: "**", min: 2}}}
	if _, _, err := resolvePolicy(bad); err == nil {
		t.Fatal("invalid rule min succeeded")
	}
}

func TestRunEarlyErrorPaths(t *testing.T) {
	if code := run([]string{"--verbose", "--dependency-scope=nope"}); code != 1 {
		t.Fatalf("invalid dependency scope exit = %d, want 1", code)
	}

	if code := run([]string{"--reusability-weight-cohesion=-1"}); code != 1 {
		t.Fatalf("invalid reusability weight exit = %d, want 1", code)
	}

	if code := run([]string{
		"--reusability-weight-cohesion=0",
		"--reusability-weight-coupling=0",
		"--reusability-weight-testability=0",
		"--reusability-weight-documentation=0",
	}); code != 1 {
		t.Fatalf("all-zero reusability weights exit = %d, want 1", code)
	}

	badProfile := filepath.Join(t.TempDir(), "missing", "cpu.prof")
	if code := run([]string{"--cpu-profile=" + badProfile}); code != 1 {
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
	original := analyze
	t.Cleanup(func() { analyze = original })

	var got reusability.ReusabilityWeights
	analyze = func(_ context.Context, cfg reusability.Config) (reusability.Report, error) {
		got = cfg.ReusabilityWeights

		return reusability.Report{}, nil
	}

	if code := run([]string{
		"--reusability-weight-cohesion=0.1",
		"--reusability-weight-coupling=0.2",
		"--reusability-weight-testability=0.3",
		"--reusability-weight-documentation=0.4",
	}); code != 0 {
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
	if code := run([]string{"--check", "./..."}); code != 2 {
		t.Fatalf("check without rules exit = %d, want 2", code)
	}
}

func TestRunCheckWithRules(t *testing.T) {
	original := analyze
	t.Cleanup(func() { analyze = original })

	analyze = func(_ context.Context, _ reusability.Config) (reusability.Report, error) {
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

	if code := run([]string{"--check", `--rule=**:0.8`}); code != 3 {
		t.Fatalf("policy violation exit = %d, want 3", code)
	}
}

func TestRunPolicySourceLogged(t *testing.T) {
	original := analyze
	t.Cleanup(func() { analyze = original })

	analyze = func(_ context.Context, _ reusability.Config) (reusability.Report, error) {
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

	if code := run([]string{"--check", `--rule=**:0.7`}); code != 0 {
		t.Fatalf("passing check exit = %d, want 0", code)
	}
}

func TestResolvePolicyValidatesRules(t *testing.T) {
	rules, _, err := resolvePolicy(ruleList{items: []ruleSpec{
		{pattern: "**/internal/**", min: 0.8},
		{pattern: "**", min: 0.6},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := policydomain.Validate(rules); err != nil {
		t.Fatal(err)
	}
}
