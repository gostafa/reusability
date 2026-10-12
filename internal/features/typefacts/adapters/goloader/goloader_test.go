// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/workerpool"
	"golang.org/x/tools/go/packages"
)

func TestTransientLoadErrorClassification(t *testing.T) {
	t.Parallel()
	for _, transient := range []error{
		syscall.EINTR, syscall.EAGAIN, syscall.EMFILE, syscall.ENFILE,
		syscall.ETIMEDOUT, syscall.ECONNRESET, syscall.ECONNABORTED,
		&net.DNSError{IsTimeout: true},
	} {
		err := fmt.Errorf("wrapped: %w", &os.PathError{Op: "load", Path: "module", Err: transient})
		if !transientLoadError(err) {
			t.Errorf("transient failure rejected: %v", err)
		}
	}
	for _, permanent := range []error{
		context.Canceled, context.DeadlineExceeded, os.ErrPermission, os.ErrNotExist,
		&net.DNSError{IsTemporary: true}, errors.New("go tool failed"),
	} {
		if transientLoadError(fmt.Errorf("wrapped: %w", permanent)) {
			t.Errorf("permanent failure retried: %v", permanent)
		}
	}
}

func TestPackageLoadRetryRecoveryAndFreshInputs(t *testing.T) {
	t.Parallel()
	deps := defaultLoaderDeps()
	cfg := &packages.Config{
		Context:    t.Context(),
		Dir:        "original",
		BuildFlags: []string{"-tags=original"},
		Env:        []string{"KEY=original"},
	}
	patterns := []string{"./original"}
	want := healthy("m/a", "a.go")
	var attempts []*packages.Config
	deps.packagesLoad = func(attempt *packages.Config, gotPatterns ...string) ([]*packages.Package, error) {
		if attempt.Context != cfg.Context || attempt.Dir != "original" ||
			attempt.BuildFlags[0] != "-tags=original" || attempt.Env[0] != "KEY=original" || gotPatterns[0] != "./original" {
			t.Errorf("retry inputs changed: %+v, %v", attempt, gotPatterns)
		}
		attempts = append(attempts, attempt)
		attempt.Dir = "mutated"
		attempt.BuildFlags[0] = "-tags=mutated"
		attempt.Env[0] = "KEY=mutated"
		gotPatterns[0] = "./mutated"
		if len(attempts) < 3 {
			return []*packages.Package{healthy("m/partial", "partial.go")}, syscall.EAGAIN
		}
		return []*packages.Package{want}, nil
	}
	got, err := retryPackageList(
		deps,
		&packageLoadRequest{config: cfg, patterns: patterns},
		loadRetryBudget,
	)
	if err != nil || len(got) != 1 || got[0] != want || len(attempts) != 3 {
		t.Fatalf("recovery = %v, %v; attempts = %d", got, err, len(attempts))
	}
	if attempts[0] == attempts[1] || attempts[1] == attempts[2] {
		t.Fatal("retry reused a mutable configuration")
	}
	if cfg.Dir != "original" || cfg.BuildFlags[0] != "-tags=original" ||
		cfg.Env[0] != "KEY=original" ||
		patterns[0] != "./original" {
		t.Fatalf("caller inputs mutated: %+v, %v", cfg, patterns)
	}
}

func TestPackageLoadRetryFailureBudgets(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		failure  error
		budget   time.Duration
		attempts int
	}{
		{name: "attempts", failure: syscall.EMFILE, budget: loadRetryBudget, attempts: 3},
		{name: "scheduling", failure: syscall.EAGAIN, budget: time.Nanosecond, attempts: 1},
		{name: "unknown", failure: errors.New("go tool failed"), budget: loadRetryBudget, attempts: 1},
		{name: "permission", failure: os.ErrPermission, budget: loadRetryBudget, attempts: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			deps := defaultLoaderDeps()
			calls := 0
			deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
				calls++
				return []*packages.Package{healthy("m/partial", "partial.go")}, test.failure
			}
			got, err := retryPackageList(
				deps,
				&packageLoadRequest{
					config:   &packages.Config{Context: t.Context()},
					patterns: []string{"./"},
				},
				test.budget,
			)
			var retryErr *backoff.RetryError
			if got != nil || !errors.Is(err, test.failure) || calls != test.attempts ||
				errors.As(err, &retryErr) {
				t.Fatalf("failure = %v, %v; calls = %d", got, err, calls)
			}
		})
	}
}

func TestPackageLoadCancellationBeforeAndDuringAttempt(t *testing.T) {
	t.Parallel()
	for _, alreadyCanceled := range []bool{false, true} {
		ctx, cancel := context.WithCancelCause(t.Context())
		sentinel := errors.New("stop loading")
		if alreadyCanceled {
			cancel(sentinel)
		}
		deps := defaultLoaderDeps()
		calls := 0
		deps.packagesLoad = func(cfg *packages.Config, _ ...string) ([]*packages.Package, error) {
			calls++
			if cfg.Context != ctx {
				t.Error("caller context was lost")
			}
			cancel(sentinel)
			return nil, syscall.EAGAIN
		}
		got, err := retryPackageList(
			deps,
			&packageLoadRequest{config: &packages.Config{Context: ctx}},
			loadRetryBudget,
		)
		cancel(sentinel)
		wantCalls := 1
		if alreadyCanceled {
			wantCalls = 0
		}
		if got != nil || !errors.Is(err, sentinel) || errors.Is(err, syscall.EAGAIN) ||
			calls != wantCalls {
			t.Fatalf("cancel = %v, %v; calls = %d", got, err, calls)
		}
	}
}

func TestPackageLoadCancellationDuringBackoff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	sentinel := errors.New("stop waiting")
	deps := defaultLoaderDeps()
	calls := 0
	var cancellation *time.Timer
	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		calls++
		cancellation = time.AfterFunc(20*time.Millisecond, func() { cancel(sentinel) })
		return nil, syscall.EAGAIN
	}
	got, err := retryPackageList(
		deps,
		&packageLoadRequest{config: &packages.Config{Context: ctx}},
		loadRetryBudget,
	)
	cancellation.Stop()
	if got != nil || !errors.Is(err, sentinel) || calls != 1 {
		t.Fatalf("backoff cancellation = %v, %v; calls = %d", got, err, calls)
	}
}

func TestPackageLoadSchedulingBudgetAllowsCompletedAttempt(t *testing.T) {
	t.Parallel()
	deps := defaultLoaderDeps()
	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		return []*packages.Package{healthy("m/a", "a.go")}, nil
	}
	got, err := retryPackageList(
		deps,
		&packageLoadRequest{config: &packages.Config{Context: t.Context()}},
		time.Nanosecond,
	)
	if err != nil || len(got) != 1 {
		t.Fatalf("completed attempt = %v, %v", got, err)
	}
}

func TestPackageLoadCancellationDiscardsSuccessfulResult(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	deps := defaultLoaderDeps()
	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		cancel()
		return []*packages.Package{healthy("m/a", "a.go")}, nil
	}
	got, err := retryPackageList(
		deps,
		&packageLoadRequest{config: &packages.Config{Context: ctx}},
		loadRetryBudget,
	)
	if got != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled result = %v, %v", got, err)
	}
}

func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/cov\n\ngo 1.22\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestExtractRichTypeShapes(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"shapes/shapes.go": `
package shapes

type Other struct{}

type Closer interface{ Close() }

type Holder struct {
	Ptr   *Other
	List  []Other
	Dict  map[Other]Other
	Fn    func(Other) Other
	Anon  struct{ Nested *Other }
	Iface interface {
		Closer
		Do(Other) Other
	}
}

func Helper() {}

type Alias = Other

const Const = 1

var _ = Const

func (h *Holder) Use() {}
`,
		"shapes/gen.go": `// Code generated by tool. DO NOT EDIT.

package shapes

type Generated struct{ X int }

func (g *Generated) Run() {}
`,
		"iface/iface.go": `
package iface

type Speaker interface {
	Speak()
}

type Talker int

func (Talker) Speak() {}
`,
	})

	mod, pkgs, err := load(context.Background(), &outbound.FactOptions{
		Directory:        dir,
		Patterns:         nil,
		IncludeGenerated: true,
		BuildTags:        []string{"ignoretag"},
	}, defaultLoaderDeps())
	if err != nil {
		t.Fatal(err)
	}
	if mod != "example.com/cov" {
		t.Fatalf("module = %q", mod)
	}

	foundGen := false
	foundHolder := false
	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Path, "/shapes") {
			continue
		}
		for _, te := range pkg.Types {
			switch te.Name {
			case "Generated":
				foundGen = true
			case "Holder":
				foundHolder = true
				if len(te.ReferencedTypeKeys) == 0 {
					t.Fatalf("Holder refs empty: %v", te.ReferencedTypeKeys)
				}
			}
		}
	}
	if !foundHolder {
		t.Fatal("Holder not extracted")
	}
	if !foundGen {
		t.Fatal("Generated type missing with IncludeGenerated=true")
	}

	_, pkgs2, err := load(context.Background(), &outbound.FactOptions{
		Directory: dir,
		Patterns:  []string{"./shapes"},
	}, defaultLoaderDeps())
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs2 {
		for _, te := range pkg.Types {
			if te.Name == "Generated" {
				t.Fatal("Generated should be skipped when IncludeGenerated=false")
			}
		}
	}

	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Path, "/iface") {
			continue
		}
		for _, te := range pkg.Types {
			if te.Name == "Talker" && len(te.Methods) != 1 {
				t.Fatalf("Talker methods = %d", len(te.Methods))
			}
		}
	}
}

func TestFieldDocumentedAndSkipPos(t *testing.T) {
	if fieldDocumented(nil, token.Pos(1)) {
		t.Fatal("empty fieldDocs should be false")
	}
	docs := []docRange{{start: 10, end: 20, documented: true}}
	if fieldDocumented(docs, token.Pos(5)) {
		t.Fatal("pos before first range should be false")
	}
	if !fieldDocumented(docs, token.Pos(15)) {
		t.Fatal("pos inside documented range should be true")
	}

	fset := token.NewFileSet()
	f := fset.AddFile("gen.go", -1, 20)
	f.SetLinesForContent([]byte("package x\n"))
	if skipPos(&skipPosRequest{
		fset:      fset,
		policy:    includeGeneratedFiles,
		generated: map[string]bool{f.Name(): true},
		pos:       token.Pos(f.Base()),
	}) {
		t.Fatal("includeGenerated=true must not skip")
	}
}

func TestSelectPackagesTruncation(t *testing.T) {
	loaded := make([]*packages.Package, 0, 12)
	for i := 0; i < 12; i++ {
		loaded = append(loaded, &packages.Package{
			PkgPath: fmt.Sprintf("m/p%d", i),
			Errors:  []packages.Error{{Msg: "boom"}},
		})
	}
	_, err := selectPackages(loaded, failOnBrokenPackages)
	if err == nil || !strings.Contains(err.Error(), "and 2 more") {
		t.Fatalf("truncation error = %v", err)
	}
}

func TestResolveBaseDirFailures(t *testing.T) {
	deps := defaultLoaderDeps()
	deps.getwd = func() (string, error) { return "", errors.New("no cwd") }
	if got := resolveBaseDir("", deps); got != "" {
		t.Fatalf("getwd failure = %q, want empty", got)
	}

	deps = defaultLoaderDeps()
	deps.absPath = func(string) (string, error) { return "", errors.New("no abs") }
	if got := resolveBaseDir("relative", deps); got != "relative" {
		t.Fatalf("abs failure = %q, want relative", got)
	}
}

func TestLoadPackagesSeamErrors(t *testing.T) {
	sentinel := errors.New("load boom")
	deps := defaultLoaderDeps()
	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		return nil, sentinel
	}
	if _, err := loadPackages(
		context.Background(),
		&outbound.FactOptions{Patterns: []string{"./"}},
		deps,
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("load error = %v", err)
	}

	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		return nil, nil
	}
	if _, err := loadPackages(
		context.Background(),
		&outbound.FactOptions{Patterns: []string{"./x"}},
		deps,
	); err == nil ||
		!strings.Contains(err.Error(), "no packages matched") {
		t.Fatalf("empty load = %v", err)
	}

	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		return []*packages.Package{{
			PkgPath: "m/broken",
			Errors:  []packages.Error{{Msg: "x"}},
		}}, nil
	}
	if _, err := loadPackages(context.Background(), &outbound.FactOptions{
		Patterns: []string{"./x"}, ContinueOnError: true,
	}, deps); err == nil || !strings.Contains(err.Error(), "no loadable packages") {
		t.Fatalf("no loadable = %v", err)
	}
}

func TestExtractAllWorkerError(t *testing.T) {
	sentinel := errors.New("extract failed")
	deps := defaultLoaderDeps()
	deps.runExtractWorkers = func(_ context.Context, _ workerpool.RunConfig) error {
		return sentinel
	}

	pkg := healthy("m/a", "a.go")
	_, err := extractAll(
		context.Background(),
		&extractJob{
			pkgs:       []*packages.Package{pkg},
			opts:       &outbound.FactOptions{},
			modulePath: "m",
		},
		deps,
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("extractAll = %v", err)
	}

	deps.packagesLoad = func(*packages.Config, ...string) ([]*packages.Package, error) {
		return []*packages.Package{healthy("m/a", "a.go")}, nil
	}
	_, _, err = load(context.Background(), &outbound.FactOptions{Patterns: []string{"./a"}}, deps)
	if !errors.Is(err, sentinel) {
		t.Fatalf("load extract error = %v", err)
	}
}

func TestWalkBodyNilBody(t *testing.T) {
	var facts domain.MethodFacts
	walkBody(&walkBodyRequest{decl: &ast.FuncDecl{}, facts: &facts})
}

func TestInvalidPositionsAndPrivateDocs(t *testing.T) {
	t.Parallel()
	if lineCount(token.NewFileSet(), token.NoPos, token.NoPos) != 0 {
		t.Fatal("invalid positions must have zero lines")
	}
	if exp, doc := countTypeDocs(
		nil,
		types.NewTypeName(token.NoPos, nil, "private", nil),
	); exp != 0 ||
		doc != 0 {
		t.Fatalf("private docs = %d/%d", exp, doc)
	}
	if relativePositionFile("", "file.go") != "file.go" ||
		relativePositionFile("/root", "/outside/file.go") != "/outside/file.go" {
		t.Fatal("outside paths must remain intact")
	}
}

func TestSelectionGuards(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset,
		"p.go",
		"package p; type T struct { X int }; func (t T) M() { t.M(); _ = T.M; _ = t.X }",
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{Selections: make(map[*ast.SelectorExpr]*types.Selection)}
	var config types.Config
	if _, err := config.Check("p", fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}
	for _, selection := range info.Selections {
		switch selection.Kind() {
		case types.MethodExpr:
			recordSelectionKind(selection, nil)
		case types.MethodVal:
			recordFieldUse(selection, nil)
			fn := selection.Obj().(*types.Func)
			recordSiblingCall(
				selection,
				&selectionRequest{
					methodIndex: map[*types.Func]int{fn: 0},
					excludeSelf: true,
					selfIdx:     0,
				},
			)
		case types.FieldVal:
			recordSiblingCall(selection, nil)
		}
	}
}

func TestIndexTypeDeclSkipsNonTypeSpec(t *testing.T) {
	decl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{{Name: "x"}}}},
	}
	idx := &syntaxIndex{typeDocs: map[types.Object]bool{}}
	indexTypeDecl(&types.Info{}, idx, decl)
	if idx.fieldDocs != nil {
		t.Fatalf("fieldDocs = %v, want nil", idx.fieldDocs)
	}
}

func TestExtractPackageSkipsNonNamedTypeName(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("b.go", -1, 20)
	file.SetLinesForContent([]byte("package m\n"))

	tpkg := types.NewPackage("m", "m")

	tn := types.NewTypeName(token.Pos(file.Base()), tpkg, "T", nil)
	_ = types.NewTypeParam(tn, types.NewInterfaceType(nil, nil))
	tpkg.Scope().Insert(tn)

	pkg := &packages.Package{
		PkgPath:   "m",
		Fset:      fset,
		Types:     tpkg,
		TypesInfo: &types.Info{},
		Syntax:    []*ast.File{{Name: &ast.Ident{Name: "m"}}},
	}
	out := extractPackage(pkg, &extractorOptions{})
	if len(out.Types) != 0 {
		t.Fatalf("types = %+v, want none", out.Types)
	}
}

func TestSortedMethodsSkipsMissingDecls(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"m.go": `
package m
type T struct{}
func (t *T) M() {}
`,
	})
	cfg := &packages.Config{
		Dir:  dir,
		Mode: loadMode,
	}
	loaded, err := packages.Load(cfg, "./")
	if err != nil || len(loaded) == 0 || loaded[0].Types == nil {
		t.Fatalf("load: %v pkgs=%d", err, len(loaded))
	}
	pkg := loaded[0]
	tn, _ := pkg.Types.Scope().Lookup("T").(*types.TypeName)
	named, _ := tn.Type().(*types.Named)

	got := sortedMethods(&extractCtx{
		pkg: pkg,
		idx: &syntaxIndex{funcDecls: map[*types.Func]*ast.FuncDecl{}},
	}, named)
	if len(got) != 0 {
		t.Fatalf("sortedMethods = %d, want 0", len(got))
	}
}

// TestCountBranchForLoops guards that every for statement contributes to the
// cyclomatic branch count, including a conditionless "for {}", matching the
// gocyclo convention that a for statement is always a decision point.
func TestCountBranchForLoops(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node ast.Node
	}{
		{"conditionless for {}", &ast.ForStmt{}},
		{"conditional for", &ast.ForStmt{Cond: &ast.Ident{Name: "ok"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var branches domain.BranchStats
			countBranch(tt.node, &branches)

			if branches.Fors != 1 {
				t.Errorf("countBranch(%s): Fors = %d, want 1", tt.name, branches.Fors)
			}
		})
	}
}

func TestCountBranchAdditionalConstructs(t *testing.T) {
	t.Parallel()

	var branches domain.BranchStats
	for _, node := range []ast.Node{
		&ast.RangeStmt{},
		&ast.CaseClause{List: []ast.Expr{&ast.Ident{Name: "value"}}},
		&ast.CommClause{Comm: &ast.SendStmt{}},
		&ast.BinaryExpr{Op: token.LAND},
	} {
		countBranch(node, &branches)
	}

	if branches.Ranges != 1 || branches.Cases != 1 || branches.SelectComms != 1 ||
		branches.LogicalOps != 1 {
		t.Fatalf("branch counts = %+v", branches)
	}
}

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
	variant := healthy("m/a", "a.go", "a_test.go")
	testBin := &packages.Package{PkgPath: "m/a.test"}
	broken := &packages.Package{PkgPath: "m/b", Errors: []packages.Error{{Msg: "boom"}}}

	got, err := selectPackages(
		[]*packages.Package{base, variant, testBin, broken},
		skipBrokenPackages,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0].PkgPath != "m/a" {
		t.Fatalf("selected %v", got)
	}

	if len(got[0].CompiledGoFiles) != 2 {
		t.Fatal("did not prefer the test-augmented variant")
	}

	if _, err := selectPackages(
		[]*packages.Package{base, broken},
		failOnBrokenPackages,
	); err == nil {
		t.Fatal("broken package without ContinueOnError should error")
	}

	missingTypes := &packages.Package{PkgPath: "m/missing-types"}
	if _, err := selectPackages(
		[]*packages.Package{missingTypes},
		failOnBrokenPackages,
	); err == nil ||
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

	if got := resolveBaseDir("", defaultLoaderDeps()); !filepath.IsAbs(got) {
		t.Errorf("empty → cwd abs, got %q", got)
	}

	if got := resolveBaseDir("relative/path", defaultLoaderDeps()); !filepath.IsAbs(got) {
		t.Errorf("relative → abs, got %q", got)
	}
}

func fixtureDir() string { return filepath.Join("..", "..", "..", "..", "..", "testdata", "fixture") }

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Black-box: the loader extracts real facts from the fixture module through
// the outbound port.
func TestLoadFixture(t *testing.T) {
	t.Parallel()

	mod, pkgs, err := New()(context.Background(), &outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./..."},
	})
	if err != nil {
		t.Fatal(err)
	}

	if mod != "example.com/fixture" {
		t.Fatalf("module = %q", mod)
	}

	var orders *domain.PackageExtract

	for i := range pkgs {
		if pkgs[i].Path == "example.com/fixture/orders" {
			orders = &pkgs[i]
		}
	}

	if orders == nil {
		t.Fatal("orders package not extracted")
	}

	if !orders.InModule {
		t.Error("orders should be in-module")
	}

	var order *domain.TypeExtract

	for i := range orders.Types {
		if orders.Types[i].Name == "Order" {
			order = &orders.Types[i]
		}
	}

	if order == nil {
		t.Fatal("Order type not extracted")
	}

	if len(order.Methods) != 3 {
		t.Errorf("Order methods = %d, want 3", len(order.Methods))
	}

	if !slices.Contains(order.ReferencedTypeKeys, "example.com/fixture/store.Store") {
		t.Errorf("Order refs = %v, want to include store.Store", order.ReferencedTypeKeys)
	}
}

func TestLoadMethodDetails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/details\n\ngo 1.23\n")
	writeFile(t, dir, "details.go", `package details

var PublicVar, privateVar int
var _ = 1

const PublicConst = 1
const privateConst = 2

type Widget struct {
	Value int
}

func init() {}

func Exported(x int) int {
	if x > 0 && x < 10 {
		return x
	}
	for i := 0; i < x; i++ {
		x--
	}
	return x
}

func helper() {}

func (w *Widget) Method(x int) int {
	switch x {
	case 1:
		return w.Value
	default:
		return 0
	}
}
`)
	writeFile(t, dir, "details_test.go", `package details

var TestVar int
const TestConst = 1

func TestHelper() {}
`)
	writeFile(t, dir, "generated.go", `// Code generated by test. DO NOT EDIT.

package details

var GeneratedVar int
const GeneratedConst = 1

func GeneratedFunc() {}
`)

	_, pkgs, err := New()(context.Background(), &outbound.FactOptions{
		Directory: dir,
		Patterns:  []string{"."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("packages = %d, want 1", len(pkgs))
	}
	pkg := pkgs[0]

	widget := findTypeExtract(t, pkg.Types, "Widget")
	method := findMethod(t, widget.Methods, "Method")
	if method.Lines != 8 {
		t.Fatalf("Method lines = %d, want 8", method.Lines)
	}
	if method.Branches.Cases != 1 {
		t.Fatalf("Method branches = %+v, want one non-default case", method.Branches)
	}
}

func findTypeExtract(t *testing.T, types []domain.TypeExtract, name string) domain.TypeExtract {
	t.Helper()

	for i := range types {
		if types[i].Name == name {
			return types[i]
		}
	}

	t.Fatalf("type %s not found", name)

	return domain.TypeExtract{}
}

func findMethod(t *testing.T, methods []domain.MethodFacts, name string) domain.MethodFacts {
	t.Helper()

	for i := range methods {
		if methods[i].Name == name {
			return methods[i]
		}
	}

	t.Fatalf("method %s not found", name)

	return domain.MethodFacts{}
}

// Black-box: a pattern matching nothing is an error.
func TestLoadNoMatch(t *testing.T) {
	t.Parallel()

	_, _, err := New()(context.Background(), &outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./does-not-exist"},
	})
	if err == nil {
		t.Fatal("expected error for a pattern matching no packages")
	}
}
