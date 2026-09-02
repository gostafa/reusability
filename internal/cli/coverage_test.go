package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

func TestOverrideListErrorsAndString(t *testing.T) {
	var overrides overrideList
	if err := overrides.Set(" types = 3.5 "); err != nil {
		t.Fatal(err)
	}
	if got := overrides.String(); got != "types=3.5" {
		t.Fatalf("String() = %q", got)
	}

	for _, value := range []string{"types", " =1", "types=not-a-number"} {
		if err := overrides.Set(value); err == nil {
			t.Errorf("Set(%q) succeeded, want error", value)
		}
	}
}

func TestResolvePolicyRequiresThresholdsAndIgnoresConfigFiles(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(config, []byte("not: valid\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := resolvePolicy(overrideList{}, overrideList{}); err == nil {
		t.Fatal("empty flag policy succeeded")
	}

	maxima := overrideList{items: []override{{key: policydomain.KeyTypes, value: 5}}}
	policy, source, err := resolvePolicy(maxima, overrideList{})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Package.Types.Max != 5 || source != "flag thresholds" {
		t.Fatalf("flag policy = %+v, source = %q", policy.Package.Types, source)
	}

	badMinimum := overrideList{items: []override{{key: "bogus", value: 1}}}
	if _, _, err := resolvePolicy(overrideList{}, badMinimum); err == nil {
		t.Fatal("unknown minimum override succeeded")
	}

	contradictory := overrideList{items: []override{{key: policydomain.KeyTypes, value: 6}}}
	if _, _, err := resolvePolicy(maxima, contradictory); err == nil {
		t.Fatal("minimum above configured maximum succeeded")
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

		return reusability.Report{
			SchemaVersion: "2",
			Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
			Module:        "example.com/m",
		}, nil
	}

	out := filepath.Join(t.TempDir(), "r.json")
	code := run([]string{
		"--format=json",
		"--output=" + out,
		"--reusability-weight-cohesion=0.1",
		"--reusability-weight-coupling=0",
		"--reusability-weight-testability=0.3",
		"--reusability-weight-documentation=0.6",
		"./...",
	})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}

	want := metrics.ReusabilityWeights{
		Cohesion:      0.1,
		Coupling:      0,
		Testability:   0.3,
		Documentation: 0.6,
	}
	if got != want {
		t.Fatalf("weights = %+v, want %+v", got, want)
	}
}

func TestResolvePolicyOverrideSource(t *testing.T) {
	maxima := overrideList{items: []override{{key: policydomain.KeyTypes, value: 20}}}
	_, source, err := resolvePolicy(maxima, overrideList{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "flag thresholds") {
		t.Fatalf("source = %q", source)
	}
}

func stubAnalyze(t *testing.T) {
	t.Helper()
	original := analyze
	t.Cleanup(func() { analyze = original })
	analyze = func(context.Context, reusability.Config) (reusability.Report, error) {
		return reusability.Report{
			SchemaVersion: "2",
			Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
			Module:        "example.com/m",
		}, nil
	}
}

func TestRunCanceledAnalysis(t *testing.T) {
	original := analyze
	t.Cleanup(func() { analyze = original })
	analyze = func(context.Context, reusability.Config) (reusability.Report, error) {
		return reusability.Report{}, context.Canceled
	}
	if code := run([]string{"./..."}); code != 130 {
		t.Fatalf("exit = %d, want 130", code)
	}
}

func TestRunMemoryProfileAndReportWriteErrors(t *testing.T) {
	stubAnalyze(t)

	badHeap := filepath.Join(t.TempDir(), "missing", "heap.prof")
	if code := run([]string{"--memory-profile=" + badHeap, "./..."}); code != 1 {
		t.Fatalf("bad memory profile exit = %d", code)
	}

	outDir := t.TempDir()
	if err := os.Chmod(outDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outDir, 0o700) })
	out := filepath.Join(outDir, "report.json")
	if code := run([]string{"--format=json", "--output=" + out, "./..."}); code != 1 {
		t.Fatalf("unwritable output exit = %d", code)
	}
}

func TestRunWebDefaultOpensBrowser(t *testing.T) {
	stubAnalyze(t)
	origTerm, origOpen := isTerminal, openBrowser
	t.Cleanup(func() { isTerminal, openBrowser = origTerm, origOpen })

	dir := t.TempDir()
	t.Chdir(dir)
	isTerminal = func() bool { return true }
	openBrowser = func(string) error { return errors.New("no browser") }

	if code := run([]string{"--format=web", "./..."}); code != 0 {
		t.Fatalf("web default exit = %d", code)
	}
	if _, err := os.Stat(filepath.Join(dir, defaultWebReportName)); err != nil {
		t.Fatalf("default web report missing: %v", err)
	}
}

func TestRunCPUStopProfileError(t *testing.T) {
	stubAnalyze(t)
	orig := startCPU
	t.Cleanup(func() { startCPU = orig })

	startCPU = func(string) (func() error, error) {
		return func() error { return errors.New("stop failed") }, nil
	}
	if code := run(
		[]string{"--cpu-profile=" + filepath.Join(t.TempDir(), "cpu.prof"), "./..."},
	); code != 0 {
		t.Fatalf("exit = %d, want 0 (stop error is logged only)", code)
	}
}

func TestRunWebHelpTerminalBrowserWarn(t *testing.T) {
	origTerm, origOpen := isTerminal, openBrowser
	t.Cleanup(func() { isTerminal, openBrowser = origTerm, origOpen })
	isTerminal = func() bool { return true }
	openBrowser = func(string) error { return errors.New("open failed") }
	if code := runWebHelp(); code != 0 {
		t.Fatalf("runWebHelp exit = %d", code)
	}
}

func TestWriteHelpDocsCloseAndWriteErrors(t *testing.T) {
	origCreate, origClose, origDocs := createHelpTemp, closeHelpFile, writeDocs
	t.Cleanup(func() {
		createHelpTemp, closeHelpFile, writeDocs = origCreate, origClose, origDocs
	})

	closeHelpFile = func(*os.File) error { return errors.New("close failed") }
	if _, err := writeHelpDocs(); err == nil {
		t.Fatal("want close error")
	}

	closeHelpFile = origClose
	writeDocs = func(outbound.Sink, string) error { return errors.New("docs failed") }
	if _, err := writeHelpDocs(); err == nil {
		t.Fatal("want docs write error")
	}
}
