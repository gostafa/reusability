package goloader

import (
	"go/types"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func healthy(path string, files ...string) *packages.Package {
	return &packages.Package{
		PkgPath:         path,
		CompiledGoFiles: files,
		Types:           types.NewPackage(path, filepath.Base(path)),
		TypesInfo:       &types.Info{},
	}
}

// White-box: dedup test variants, drop synthesized test binaries, apply the
// error policy.
func TestSelectPackages(t *testing.T) {
	t.Parallel()

	base := healthy("m/a", "a.go")
	variant := healthy("m/a", "a.go", "a_test.go") // more files → preferred
	testBin := &packages.Package{PkgPath: "m/a.test"}
	broken := &packages.Package{PkgPath: "m/b", Errors: []packages.Error{{Msg: "boom"}}}

	got, err := selectPackages([]*packages.Package{base, variant, testBin, broken}, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0].PkgPath != "m/a" {
		t.Fatalf("selected %v", got)
	}

	if len(got[0].CompiledGoFiles) != 2 {
		t.Fatal("did not prefer the test-augmented variant")
	}

	if _, err := selectPackages([]*packages.Package{base, broken}, false); err == nil {
		t.Fatal("broken package without ContinueOnError should error")
	}

	missingTypes := &packages.Package{PkgPath: "m/missing-types"}
	if _, err := selectPackages([]*packages.Package{missingTypes}, false); err == nil ||
		!strings.Contains(err.Error(), "type information unavailable") {
		t.Fatalf("missing type information error = %v", err)
	}
}

// White-box: the main module wins; otherwise any module; otherwise empty.
func TestMainModulePath(t *testing.T) {
	t.Parallel()

	main := &packages.Package{Module: &packages.Module{Path: "example.com/main", Main: true}}

	dep := &packages.Package{Module: &packages.Module{Path: "example.com/dep"}}
	if got := mainModulePath([]*packages.Package{dep, main}); got != "example.com/main" {
		t.Errorf("got %q, want main", got)
	}

	if got := mainModulePath([]*packages.Package{dep}); got != "example.com/dep" {
		t.Errorf("got %q, want dep fallback", got)
	}

	if got := mainModulePath([]*packages.Package{{}}); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// White-box: base dir is always absolutized.
func TestResolveBaseDir(t *testing.T) {
	t.Parallel()

	if got := resolveBaseDir(""); !filepath.IsAbs(got) {
		t.Errorf("empty → cwd abs, got %q", got)
	}

	if got := resolveBaseDir("relative/path"); !filepath.IsAbs(got) {
		t.Errorf("relative → abs, got %q", got)
	}
}
