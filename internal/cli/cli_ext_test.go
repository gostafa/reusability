package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/cli"
)

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

	if code := cli.Run([]string{
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

	if got["schema_version"] != "4" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkgs := got["packages"].([]any)
	if len(pkgs) < 7 {
		t.Errorf("packages = %d, want >= 7", len(pkgs))
	}

	// Schema v4 structural facts are present on packages and types.
	first := pkgs[0].(map[string]any)
	for _, key := range []string{"afferent", "efferent", "funcs", "vars", "consts", "functions", "variables", "constants"} {
		if _, ok := first[key]; !ok {
			t.Errorf("package is missing structural fact %q", key)
		}
	}

	if types := first["types"].([]any); len(types) > 0 {
		typ := types[0].(map[string]any)
		for _, key := range []string{"fields", "methods", "kind", "position", "field_details", "method_details"} {
			if _, ok := typ[key]; !ok {
				t.Errorf("type is missing structural fact %q", key)
			}
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

	if code := cli.Run([]string{"--web", "--output=" + out, "./..."}); code != 0 {
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

	if code := cli.Run([]string{"--web", "--format=json"}); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

// Black-box: --version succeeds.
func TestRunVersion(t *testing.T) {
	t.Parallel()

	if code := cli.Run([]string{"--version"}); code != 0 {
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
			t.Setenv("TMPDIR", tmp) // darwin/linux
			t.Setenv("TMP", tmp)    // windows

			if code := cli.Run(args); code != 0 {
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

	if code := cli.Run([]string{"--help"}); code != 2 {
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

// Black-box: a violated condition exits 3. `types` max 0 is broken by any
// package that declares a type, so this does not depend on fixture metrics.
func TestRunCheckFailsExitsThree(t *testing.T) {
	chdirFixture(t)

	if code := cli.Run(
		[]string{"--max", "types=0", "--output", filepath.Join(t.TempDir(), "r.txt"), "./..."},
	); code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
}

// Black-box: a satisfiable flag-only policy passes with exit 0, exercising
// structural, package-metric, and type-metric bounds at once.
func TestRunCheckFlagPolicyPasses(t *testing.T) {
	chdirFixture(t)

	args := []string{
		"--output", filepath.Join(t.TempDir(), "r.txt"),
		"--max=types=100000",
		"--max=funcs=100000",
		"--max=vars=100000",
		"--max=consts=100000",
		"--max=afferent=100000",
		"--max=efferent=100000",
		"--max=fields=100000",
		"--max=methods=100000",
		"--min=type.reusability=0",
		"./...",
	}

	if code := cli.Run(args); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

// Black-box: an unknown override key is a usage error (exit 2), and gating
// never runs without a policy flag.
func TestRunCheckKeyAndTriggers(t *testing.T) {
	chdirFixture(t)

	out := filepath.Join(t.TempDir(), "r.txt")

	if code := cli.Run([]string{"--max", "bogus=5", "--output", out, "./..."}); code != 2 {
		t.Fatalf("unknown key exit = %d, want 2", code)
	}

	if code := cli.Run([]string{"--check", "--output", out, "./..."}); code != 2 {
		t.Fatalf("empty check exit = %d, want 2", code)
	}

	// No policy flag → no gate, even though types=0 would fail under one.
	if code := cli.Run([]string{"--output", out, "./..."}); code != 0 {
		t.Fatalf("ungated exit = %d, want 0", code)
	}
}
