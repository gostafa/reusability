// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/bitset"
	"github.com/gostafa/reusability/internal/shared/workerpool"
	"golang.org/x/tools/go/packages"
)

var (
	errNoPackagesMatched   = errors.New("no packages matched patterns")
	errNoLoadablePackages  = errors.New("no loadable packages matched patterns")
	errPackageLoadFailures = errors.New("package load errors (use ContinueOnError to skip)")
)

// Load loads the requested patterns once, deduplicates test variants, and
// extracts per-package facts with bounded package-level concurrency.
func (loader *Loader) Load(
	ctx context.Context,
	opts *outbound.FactOptions,
) (modulePath string, extracts []domain.PackageExtract, err error) {
	modulePath, extracts, err = loader.extract(ctx, opts)
	if err != nil {
		return emptyString, nil, fmt.Errorf("loader extract: %w", err)
	}

	return modulePath, extracts, nil
}

// New returns a compiler-backed fact source.
func New() *Loader { return &Loader{extract: extractFacts} }

func extractFacts(
	ctx context.Context,
	opts *outbound.FactOptions,
) (modulePath string, extracts []domain.PackageExtract, err error) {
	modulePath, extracts, err = load(ctx, opts, defaultLoaderDeps())
	if err != nil {
		return emptyString, nil, fmt.Errorf("goloader load: %w", err)
	}

	return modulePath, extracts, nil
}

func defaultLoaderDeps() *loaderDeps {
	return &loaderDeps{
		packagesLoad:      packages.Load,
		runExtractWorkers: workerpool.Run,
		getwd:             os.Getwd,
		absPath:           filepath.Abs,
	}
}

func addInterfaceRefs(refs *refCollector, iface *types.Interface) {
	for etyp := range iface.EmbeddedTypes() {
		refs.addType(etyp)
	}

	for method := range iface.ExplicitMethods() {
		if sig, ok := method.Type().(*types.Signature); ok {
			addSignatureRefs(refs, sig)
		}
	}
}

func addSignatureRefs(refs *refCollector, sig *types.Signature) {
	for param := range sig.Params().Variables() {
		refs.addType(param.Type())
	}

	for result := range sig.Results().Variables() {
		refs.addType(result.Type())
	}
}

func addStructRefs(refs *refCollector, structType *types.Struct) {
	for field := range structType.Fields() {
		refs.addType(field.Type())
	}
}

func addTypeArgRefs(refs *refCollector, named *types.Named) {
	args := named.TypeArgs()

	for typeArg := range args.Types() {
		refs.addType(typeArg)
	}
}

func countBranch(n ast.Node, branches *domain.BranchStats) {
	if countCtrlBranch(n, branches) {
		return
	}

	countExprBranch(n, branches)
}

func countCtrlBranch(n ast.Node, branches *domain.BranchStats) bool {
	switch n.(type) {
	case *ast.IfStmt:
		branches.Ifs++
	case *ast.ForStmt:
		branches.Fors++
	case *ast.RangeStmt:
		branches.Ranges++
	default:
		return false
	}

	return true
}

func countExprBranch(n ast.Node, branches *domain.BranchStats) {
	switch node := n.(type) {
	case *ast.CaseClause:
		countCaseClause(node, branches)
	case *ast.CommClause:
		countCommClause(node, branches)
	case *ast.BinaryExpr:
		countLogicalOp(node, branches)
	default:
	}
}

func countCaseClause(n *ast.CaseClause, branches *domain.BranchStats) {
	if n.List != nil {
		branches.Cases++
	}
}

func countCommClause(n *ast.CommClause, branches *domain.BranchStats) {
	if n.Comm != nil {
		branches.SelectComms++
	}
}

func countLogicalOp(n *ast.BinaryExpr, branches *domain.BranchStats) {
	if n.Op == token.LAND || n.Op == token.LOR {
		branches.LogicalOps++
	}
}

func descendRef(refs *refCollector, typ types.Type) {
	if descendContainer(refs, typ) {
		return
	}

	descendComposite(refs, typ)
}

func descendContainer(refs *refCollector, typ types.Type) bool {
	switch concrete := typ.(type) {
	case *types.Map:
		refs.addType(concrete.Key())
		refs.addType(concrete.Elem())

		return true
	case interface{ Elem() types.Type }:
		refs.addType(concrete.Elem())

		return true
	default:
		return false
	}
}

func descendComposite(refs *refCollector, typ types.Type) {
	switch concrete := typ.(type) {
	case *types.Signature:
		addSignatureRefs(refs, concrete)
	case *types.Struct:
		addStructRefs(refs, concrete)
	case *types.Interface:
		addInterfaceRefs(refs, concrete)
	default:
	}
}

func extractAll(
	ctx context.Context,
	job *extractJob,
	deps *loaderDeps,
) ([]domain.PackageExtract, error) {
	// extractAll.
	state := newExtractWorkerState(job, deps)
	workers := workerpool.Workers(job.opts.Workers, len(job.pkgs))

	err := deps.runExtractWorkers(ctx, workerpool.RunConfig{
		Workers:   workers,
		TaskCount: len(job.pkgs),
		Fn:        state.worker(),
	})
	if err != nil {
		return nil, fmt.Errorf("extract packages: %w", err)
	}

	return state.extracts, nil
}

func newExtractWorkerState(job *extractJob, deps *loaderDeps) *extractWorkerState {
	return &extractWorkerState{
		job:      job,
		extracts: make([]domain.PackageExtract, len(job.pkgs)),
		opts:     newExtractorOptions(job, deps),
	}
}

func newExtractorOptions(job *extractJob, deps *loaderDeps) extractorOptions {
	return extractorOptions{
		analyzed:   analyzedPkgPaths(job.pkgs),
		modulePath: job.modulePath,
		baseDir:    resolveBaseDir(job.opts.Directory, deps),
		generated:  generatedPolicy(job.opts),
	}
}

func analyzedPkgPaths(pkgs []*packages.Package) map[string]bool {
	analyzed := make(map[string]bool, len(pkgs))

	for i := range pkgs {
		analyzed[pkgs[i].PkgPath] = true
	}

	return analyzed
}

func generatedPolicy(opts *outbound.FactOptions) generatedInclusion {
	if opts.IncludeGenerated {
		return includeGeneratedFiles
	}

	return excludeGeneratedFiles
}

func (s *extractWorkerState) worker() func(int) error {
	return func(i int) error {
		pkg := s.job.pkgs[i]

		s.extracts[i] = extractPackage(pkg, &s.opts)
		pkg.Syntax = nil
		pkg.TypesInfo = nil
		pkg.Types = nil

		return nil
	}
}

func extractPackage(pkg *packages.Package, opts *extractorOptions) domain.PackageExtract {
	idx := indexSyntax(pkg)
	ctx := newExtractCtx(pkg, opts, &idx)
	out := domain.PackageExtract{
		Path:     pkg.PkgPath,
		InModule: inModule(pkg, opts.modulePath),
		Imports:  importPaths(pkg),
	}
	collectTypeExtracts(&out, ctx, pkg.Types.Scope())

	return out
}

func newExtractCtx(pkg *packages.Package, opts *extractorOptions, idx *syntaxIndex) *extractCtx {
	return &extractCtx{
		pkg:              pkg,
		idx:              idx,
		analyzed:         opts.analyzed,
		baseDir:          opts.baseDir,
		includeGenerated: opts.generated == includeGeneratedFiles,
	}
}

func collectTypeExtracts(out *domain.PackageExtract, ctx *extractCtx, scope *types.Scope) {
	names := scope.Names()

	for i := range names {
		appendNamedType(out, ctx, scope, names[i])
	}
}

func appendNamedType(out *domain.PackageExtract, ctx *extractCtx, scope *types.Scope, name string) {
	typeName, named, ok := lookupNamedType(scope, name)

	if !ok {
		return
	}

	if skipCtxPos(ctx, typeName.Pos()) {
		return
	}

	out.Types = append(out.Types, extractType(ctx, typeName, named))
}

func lookupNamedType(scope *types.Scope, name string) (*types.TypeName, *types.Named, bool) {
	typeName, ok := scope.Lookup(name).(*types.TypeName)

	if !ok || typeName.IsAlias() {
		return nil, nil, false
	}

	named, ok := typeName.Type().(*types.Named)

	if !ok {
		return nil, nil, false
	}

	return typeName, named, true
}

func extractType(ctx *extractCtx, typeName *types.TypeName, named *types.Named) domain.TypeExtract {
	out := baseTypeExtract(ctx, typeName, named)
	refs := newRefCollector(typeName, ctx.analyzed)
	collected := structFields(named, refs)
	fields, fieldIndex, fieldPositions := collected.fields, collected.fieldIndex, collected.positions

	out.Fields = fields

	docMethods := fillTypeMethods(&out, ctx, named, refs, fieldIndex)
	attachTypeRefsDocs(&out, ctx, typeName, refs, fieldPositions, docMethods)

	return out
}

func attachTypeRefsDocs(
	out *domain.TypeExtract,
	ctx *extractCtx,
	typeName *types.TypeName,
	refs *refCollector,
	fieldPositions []token.Pos,
	docMethods []methodDocInput,
) {
	out.ReferencedTypeKeys = sortedRefKeys(refs.seen)
	out.ExportedMembers, out.DocumentedExportedMembers = memberDocs(
		ctx.idx, typeName, out.Fields, fieldPositions, docMethods,
	)
}

func baseTypeExtract(
	ctx *extractCtx,
	typeName *types.TypeName,
	named *types.Named,
) domain.TypeExtract {
	return domain.TypeExtract{
		Name:     typeName.Name(),
		Exported: typeName.Exported(),
		Kind:     typeKind(named),
		Pos:      position(ctx.pkg.Fset, ctx.baseDir, typeName.Pos()),
	}
}

func fillTypeMethods(
	out *domain.TypeExtract,
	ctx *extractCtx,
	named *types.Named,
	refs *refCollector,
	fieldIndex map[*types.Var]int,
) []methodDocInput {
	methods := sortedMethods(ctx, named)
	methodIndex := indexMethodDecls(methods)
	docs := initializeMethodFacts(out, methods)

	for i := range methods {
		facts, doc := methodFacts(ctx, methods[i], refs, fieldIndex, methodIndex, len(out.Fields))

		out.Methods = append(out.Methods, facts)
		docs = append(docs, doc)
	}

	return docs
}

func initializeMethodFacts(out *domain.TypeExtract, methods []methodDecl) []methodDocInput {
	out.Methods = make([]domain.MethodFacts, zero, len(methods))

	return make([]methodDocInput, zero, len(methods))
}

func indexMethodDecls(methods []methodDecl) map[*types.Func]int {
	methodIndex := make(map[*types.Func]int, len(methods))

	for i := range methods {
		methodIndex[methods[i].fn] = i
	}

	return methodIndex
}

func fieldDocumented(fieldDocs []docRange, pos token.Pos) bool {
	i := sort.Search(len(fieldDocs), func(i int) bool { return fieldDocs[i].start > pos })

	if i == zero {
		return false
	}

	r := fieldDocs[i-one]

	return pos >= r.start && pos <= r.end && r.documented
}

func importPaths(pkg *packages.Package) []string {
	if len(pkg.Imports) == zero {
		return nil
	}

	paths := make([]string, zero, len(pkg.Imports))

	for path := range pkg.Imports {
		paths = append(paths, path)
	}

	slices.Sort(paths)

	return paths
}

func inModule(pkg *packages.Package, modulePath string) bool {
	return pkg.Module != nil && modulePath != emptyString && pkg.Module.Path == modulePath
}

func indexSyntax(pkg *packages.Package) syntaxIndex {
	idx := syntaxIndex{
		generated: make(map[string]bool),
		funcDecls: make(map[*types.Func]*ast.FuncDecl),
		typeDocs:  make(map[types.Object]bool),
	}

	for i := range pkg.Syntax {
		indexFile(pkg, pkg.Syntax[i], &idx)
	}

	slices.SortFunc(idx.fieldDocs, func(a, b docRange) int {
		return cmp.Compare(a.start, b.start)
	})

	return idx
}

func indexFile(pkg *packages.Package, file *ast.File, idx *syntaxIndex) {
	filename := pkg.Fset.Position(file.Package).Filename

	idx.generated[filename] = ast.IsGenerated(file)

	for i := range file.Decls {
		indexDecl(pkg.TypesInfo, idx, file.Decls[i])
	}
}

func indexDecl(info *types.Info, idx *syntaxIndex, decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		indexFuncDecl(info, idx, d)
	case *ast.GenDecl:
		indexTypeDecl(info, idx, d)
	default:
	}
}

func indexFuncDecl(info *types.Info, idx *syntaxIndex, decl *ast.FuncDecl) {
	if decl.Recv == nil {
		return
	}

	if fn, ok := info.Defs[decl.Name].(*types.Func); ok {
		idx.funcDecls[fn] = decl
	}
}

func indexTypeDecl(info *types.Info, idx *syntaxIndex, decl *ast.GenDecl) {
	if decl.Tok != token.TYPE {
		return
	}

	ctx := typeSpecIndex{info: info, idx: idx, decl: decl}

	for i := range decl.Specs {
		indexTypeSpec(ctx, decl.Specs[i])
	}
}

func indexTypeSpec(ctx typeSpecIndex, spec ast.Spec) {
	typeSpec, ok := spec.(*ast.TypeSpec)

	if !ok {
		return
	}

	recordTypeDoc(ctx, typeSpec)
	appendStructFieldDocs(ctx.idx, typeSpec)
}

func recordTypeDoc(ctx typeSpecIndex, spec *ast.TypeSpec) {
	documented := spec.Doc != nil || (len(ctx.decl.Specs) == one && ctx.decl.Doc != nil)

	if obj := ctx.info.Defs[spec.Name]; obj != nil {
		ctx.idx.typeDocs[obj] = documented
	}
}

func appendStructFieldDocs(idx *syntaxIndex, spec *ast.TypeSpec) {
	structType, ok := spec.Type.(*ast.StructType)

	if !ok || structType.Fields == nil {
		return
	}

	for i := range structType.Fields.List {
		field := structType.Fields.List[i]

		idx.fieldDocs = append(idx.fieldDocs, docRange{
			start:      field.Pos(),
			end:        field.End(),
			documented: field.Doc != nil || field.Comment != nil,
		})
	}
}

func lineCount(fset *token.FileSet, start, end token.Pos) int {
	startLine := fset.Position(start).Line
	endLine := fset.Position(end).Line

	if startLine <= zero || endLine < startLine {
		return zero
	}

	return endLine - startLine + one
}

func load(
	ctx context.Context,
	opts *outbound.FactOptions,
	deps *loaderDeps,
) (modulePath string, extracts []domain.PackageExtract, err error) {
	// load.
	pkgs, err := loadPackages(ctx, opts, deps)
	if err != nil {
		return emptyString, nil, fmt.Errorf("goloader packages: %w", err)
	}

	modulePath, extracts, finishErr := extractLoaded(ctx, opts, deps, pkgs)
	if finishErr != nil {
		return emptyString, nil, fmt.Errorf("goloader finish: %w", finishErr)
	}

	return modulePath, extracts, nil
}

func extractLoaded(
	ctx context.Context,
	opts *outbound.FactOptions,
	deps *loaderDeps,
	pkgs []*packages.Package,
) (modulePath string, extracts []domain.PackageExtract, err error) {
	// extractLoaded.
	modulePath = mainModulePath(pkgs)

	extracts, err = extractAll(
		ctx,
		&extractJob{pkgs: pkgs, opts: opts, modulePath: modulePath},
		deps,
	)
	if err != nil {
		return emptyString, nil, fmt.Errorf("goloader extract: %w", err)
	}

	return modulePath, extracts, nil
}

func loadPackages(
	ctx context.Context,
	opts *outbound.FactOptions,
	deps *loaderDeps,
) ([]*packages.Package, error) {
	// loadPackages.
	loaded, patterns, err := invokePackagesLoad(ctx, opts, deps)
	if err != nil {
		return nil, fmt.Errorf(errFmtInvokePackagesLoad, err)
	}

	selected, selectErr := selectLoadedPackages(
		&selectLoadedRequest{loaded: loaded, opts: opts, patterns: patterns},
	)
	if selectErr != nil {
		return nil, fmt.Errorf("select loaded packages: %w", selectErr)
	}

	return selected, nil
}

func invokePackagesLoad(
	ctx context.Context,
	opts *outbound.FactOptions,
	deps *loaderDeps,
) ([]*packages.Package, []string, error) {
	// invokePackagesLoad.
	loaded, patterns, err := loadPackageList(
		deps, packageLoadConfig(ctx, opts), packagePatterns(opts),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(errFmtInvokePackagesLoad, err)
	}

	return loaded, patterns, nil
}

func loadPackageList(
	deps *loaderDeps,
	cfg *packages.Config,
	patterns []string,
) ([]*packages.Package, []string, error) {
	// loadPackageList.
	loaded, err := deps.packagesLoad(cfg, patterns...)
	if err != nil {
		return nil, nil, fmt.Errorf("load packages: %w", err)
	}

	if len(loaded) == zero {
		return nil, nil, fmt.Errorf(errWrapWithValue, errNoPackagesMatched, patterns)
	}

	return loaded, patterns, nil
}

func packagePatterns(opts *outbound.FactOptions) []string {
	if len(opts.Patterns) == zero {
		return []string{defaultPackagePattern}
	}

	return opts.Patterns
}

func packageLoadConfig(ctx context.Context, opts *outbound.FactOptions) *packages.Config {
	cfg := &packages.Config{
		Context: ctx,
		Dir:     opts.Directory,
		Mode:    loadMode,
		Tests:   opts.IncludeTests,
	}

	if len(opts.BuildTags) > zero {
		cfg.BuildFlags = []string{buildTagsPrefix + strings.Join(opts.BuildTags, ",")}
	}

	return cfg
}

func selectLoadedPackages(req *selectLoadedRequest) ([]*packages.Package, error) {
	pkgs, err := selectPackages(req.loaded, packagePolicy(req.opts))
	if err != nil {
		return nil, fmt.Errorf("select packages: %w", err)
	}

	if len(pkgs) == zero {
		return nil, fmt.Errorf(errWrapWithValue, errNoLoadablePackages, req.patterns)
	}

	return pkgs, nil
}

func packagePolicy(opts *outbound.FactOptions) packageErrorPolicy {
	if opts.ContinueOnError {
		return skipBrokenPackages
	}

	return failOnBrokenPackages
}

func mainModulePath(pkgs []*packages.Package) string {
	if path := findMainModulePath(pkgs); path != emptyString {
		return path
	}

	return findAnyModulePath(pkgs)
}

func findMainModulePath(pkgs []*packages.Package) string {
	for i := range pkgs {
		pkg := pkgs[i]

		if pkg.Module != nil && pkg.Module.Main {
			return pkg.Module.Path
		}
	}

	return emptyString
}

func findAnyModulePath(pkgs []*packages.Package) string {
	for i := range pkgs {
		pkg := pkgs[i]

		if pkg.Module != nil {
			return pkg.Module.Path
		}
	}

	return emptyString
}

func memberDocs(
	idx *syntaxIndex,
	typeName *types.TypeName,
	fields []domain.FieldFacts,
	fieldPositions []token.Pos,
	methods []methodDocInput,
) (exported, documented int) {
	exported, documented = countTypeDocs(idx.typeDocs, typeName)
	exported, documented = addFieldDocs(idx.fieldDocs, fields, fieldPositions, exported, documented)
	exported, documented = addMethodDocs(methods, exported, documented)

	return exported, documented
}

func countTypeDocs(
	typeDocs map[types.Object]bool,
	typeName *types.TypeName,
) (exported, documented int) {
	if !typeName.Exported() {
		return zero, zero
	}

	exported = one

	if typeDocs[typeName] {
		documented = one
	}

	return exported, documented
}

func addFieldDocs(
	fieldDocs []docRange,
	fields []domain.FieldFacts,
	fieldPositions []token.Pos,
	exported, documented int,
) (exp, doc int) {
	exp, doc = exported, documented

	for i := range fields {
		if !fields[i].Exported {
			continue
		}

		exp++

		if fieldDocumented(fieldDocs, fieldPositions[i]) {
			doc++
		}
	}

	return exp, doc
}

func addMethodDocs(methods []methodDocInput, exported, documented int) (exp, doc int) {
	exp, doc = exported, documented

	for i := range methods {
		if !methods[i].exported {
			continue
		}

		exp++

		if methods[i].documented {
			doc++
		}
	}

	return exp, doc
}

func methodFacts(
	ctx *extractCtx,
	method methodDecl,
	refs *refCollector,
	fieldIndex map[*types.Var]int,
	methodIndex map[*types.Func]int,
	fieldCount int,
) (domain.MethodFacts, methodDocInput) {
	facts := newMethodFacts(ctx, method)

	if sig, ok := method.fn.Type().(*types.Signature); ok {
		refs.addType(sig)
	}

	walkBody(&walkBodyRequest{
		info: ctx.pkg.TypesInfo, decl: method.decl, self: method.fn,
		fieldCount: fieldCount, fieldIndex: fieldIndex,
		methodIndex: methodIndex, facts: &facts,
	})

	return facts, methodDocInput{exported: method.fn.Exported(), documented: method.decl.Doc != nil}
}

func newMethodFacts(ctx *extractCtx, method methodDecl) domain.MethodFacts {
	return domain.MethodFacts{
		Name:     method.fn.Name(),
		Exported: method.fn.Exported(),
		Pos:      position(ctx.pkg.Fset, ctx.baseDir, method.decl.Pos()),
		Lines:    lineCount(ctx.pkg.Fset, method.decl.Pos(), method.decl.End()),
	}
}

func newRefCollector(self *types.TypeName, analyzed map[string]bool) *refCollector {
	return &refCollector{
		self:     self,
		analyzed: analyzed,
		seen:     make(map[string]bool),
		visited:  make(map[types.Type]bool),
	}
}

func position(fset *token.FileSet, baseDir string, pos token.Pos) domain.Position {
	p := fset.Position(pos)

	return domain.Position{
		File:   relativePositionFile(baseDir, p.Filename),
		Line:   p.Line,
		Column: p.Column,
	}
}

func relativePositionFile(baseDir, file string) string {
	if baseDir == emptyString {
		return file
	}

	rel, err := filepath.Rel(baseDir, file)

	if err != nil || strings.HasPrefix(rel, parentDir) {
		return file
	}

	return filepath.ToSlash(rel)
}

func recordNamedRef(req *namedRefRequest) {
	tn := req.t.Origin().Obj()

	if tn != req.self && tn.Pkg() != nil && req.analyzed[tn.Pkg().Path()] {
		req.seen[domain.TypeKey(tn.Pkg().Path(), tn.Name())] = true
	}
}

func recordSelection(req *selectionRequest) {
	sel, ok := req.info.Selections[req.n]

	if !ok {
		return
	}

	recordSelectionKind(sel, req)
}

func recordSelectionKind(sel *types.Selection, req *selectionRequest) {
	switch sel.Kind() {
	case types.FieldVal:
		recordFieldUse(sel, req)
	case types.MethodVal:
		recordSiblingCall(sel, req)
	case types.MethodExpr:
		return
	default:
	}
}

func recordFieldUse(sel *types.Selection, req *selectionRequest) {
	field, ok := sel.Obj().(*types.Var)

	if !ok {
		return
	}

	idx, found := req.fieldIndex[field.Origin()]

	if found {
		req.facts.FieldsUsed.Set(idx)
	}
}

func recordSiblingCall(sel *types.Selection, req *selectionRequest) {
	method, ok := sel.Obj().(*types.Func)

	if !ok {
		return
	}

	idx, found := req.methodIndex[method.Origin()]

	if !found {
		return
	}

	if req.excludeSelf && idx == req.selfIdx {
		return
	}

	req.siblings[idx] = true
}

func (refs *refCollector) addType(typ types.Type) {
	unaliased := types.Unalias(typ)

	if refs.visited[unaliased] {
		return
	}

	refs.visited[unaliased] = true

	if named, ok := unaliased.(*types.Named); ok {
		recordNamedRef(&namedRefRequest{
			seen: refs.seen, self: refs.self, analyzed: refs.analyzed, t: named,
		})
		addTypeArgRefs(refs, named)

		return
	}

	descendRef(refs, unaliased)
}

func resolveBaseDir(dir string, deps *loaderDeps) string {
	if dir == emptyString {
		return resolveWorkingDir(deps)
	}

	return resolveAbsDir(dir, deps)
}

func resolveWorkingDir(deps *loaderDeps) string {
	wd, err := deps.getwd()
	if err != nil {
		return emptyString
	}

	return wd
}

func resolveAbsDir(dir string, deps *loaderDeps) string {
	abs, err := deps.absPath(dir)
	if err != nil {
		return dir
	}

	return abs
}

func selectPackages(
	loaded []*packages.Package,
	policy packageErrorPolicy,
) ([]*packages.Package, error) {
	// selectPackages.
	byPath, order := dedupePackages(loaded)
	pkgs, errs := filterPackages(
		&filterPackagesRequest{byPath: byPath, order: order, policy: policy},
	)

	err := packageSelectionError(errs)
	if err != nil {
		return nil, fmt.Errorf("package selection: %w", err)
	}

	return pkgs, nil
}

func dedupePackages(
	loaded []*packages.Package,
) (byPath map[string]*packages.Package, order []string) {
	// dedupePackages.
	byPath = make(map[string]*packages.Package, len(loaded))
	order = make([]string, zero, len(loaded))

	for i := range loaded {
		mergePackage(byPath, &order, loaded[i])
	}

	return byPath, order
}

func mergePackage(byPath map[string]*packages.Package, order *[]string, pkg *packages.Package) {
	if strings.HasSuffix(pkg.PkgPath, testPackageSuffix) {
		return
	}

	existing, ok := byPath[pkg.PkgPath]

	if !ok {
		byPath[pkg.PkgPath] = pkg
		*order = append(*order, pkg.PkgPath)

		return
	}

	if len(pkg.CompiledGoFiles) > len(existing.CompiledGoFiles) {
		byPath[pkg.PkgPath] = pkg
	}
}

func filterPackages(req *filterPackagesRequest) (pkgs []*packages.Package, errs []string) {
	state := packageFilterState{
		pkgs: make([]*packages.Package, zero, len(req.order)),
	}

	for i := range req.order {
		path := req.order[i]
		considerPackage(&state, req.byPath[path], packageConsider{path: path, policy: req.policy})
	}

	return state.pkgs, state.errs
}

func considerPackage(state *packageFilterState, pkg *packages.Package, consider packageConsider) {
	if !packageBroken(pkg) {
		state.pkgs = append(state.pkgs, pkg)

		return
	}

	if consider.policy == skipBrokenPackages {
		return
	}

	state.errs = appendPackageErrors(state.errs, consider.path, pkg)
}

func packageBroken(pkg *packages.Package) bool {
	return len(pkg.Errors) > zero || pkg.Types == nil || pkg.TypesInfo == nil
}

func appendPackageErrors(errs []string, path string, pkg *packages.Package) []string {
	out := errs

	for i := range pkg.Errors {
		out = append(out, fmt.Sprintf("%s: %s", path, pkg.Errors[i].Msg))
	}

	if len(pkg.Errors) == zero {
		out = append(out, path+": type information unavailable")
	}

	return out
}

func packageSelectionError(errs []string) error {
	if len(errs) == zero {
		return nil
	}

	shown, suffix := truncateErrors(errs)

	return fmt.Errorf(
		"%w:\n%s%s",
		errPackageLoadFailures,
		strings.Join(shown, "\n"),
		suffix,
	)
}

func truncateErrors(errs []string) (shown []string, suffix string) {
	if len(errs) <= maxShownErrors {
		return errs, emptyString
	}

	return errs[:maxShownErrors], fmt.Sprintf("\n… and %d more", len(errs)-maxShownErrors)
}

func skipCtxPos(ctx *extractCtx, pos token.Pos) bool {
	policy := excludeGeneratedFiles

	if ctx.includeGenerated {
		policy = includeGeneratedFiles
	}

	return skipPos(&skipPosRequest{
		fset: ctx.pkg.Fset, policy: policy,
		generated: ctx.idx.generated, pos: pos,
	})
}

func skipPos(req *skipPosRequest) bool {
	if req.policy == includeGeneratedFiles {
		return false
	}

	return req.generated[req.fset.Position(req.pos).Filename]
}

func sortedMethods(ctx *extractCtx, named *types.Named) []methodDecl {
	methods := collectMethodDecls(ctx, named)
	slices.SortFunc(methods, func(a, b methodDecl) int {
		return strings.Compare(a.fn.Name(), b.fn.Name())
	})

	return methods
}

func collectMethodDecls(ctx *extractCtx, named *types.Named) []methodDecl {
	methods := make([]methodDecl, zero, named.NumMethods())

	for method := range named.Methods() {
		methods = appendMethodDecl(methods, ctx, method)
	}

	return methods
}

func appendMethodDecl(
	methods []methodDecl,
	ctx *extractCtx,
	method *types.Func,
) []methodDecl {
	// appendMethodDecl.
	decl, ok := ctx.idx.funcDecls[method]

	if !ok || skipCtxPos(ctx, decl.Pos()) {
		return methods
	}

	return append(methods, methodDecl{fn: method, decl: decl})
}

func sortedRefKeys(seen map[string]bool) []string {
	if len(seen) == zero {
		return nil
	}

	keys := make([]string, zero, len(seen))

	for key := range seen {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

func structFields(named *types.Named, refs *refCollector) structFieldsResult {
	structType, ok := named.Underlying().(*types.Struct)

	if !ok {
		return structFieldsResult{}
	}

	return collectStructFields(structType, refs)
}

func collectStructFields(structType *types.Struct, refs *refCollector) structFieldsResult {
	fields := make([]domain.FieldFacts, zero, structType.NumFields())
	fieldIndex := make(map[*types.Var]int, structType.NumFields())
	positions := make([]token.Pos, zero, structType.NumFields())

	for i := zero; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		fieldIndex[field] = i
		positions = append(positions, field.Pos())
		fields = append(fields, domain.FieldFacts{
			Name:     field.Name(),
			Exported: field.Exported(),
			Embedded: field.Anonymous(),
		})
		refs.addType(field.Type())
	}

	return structFieldsResult{fields: fields, fieldIndex: fieldIndex, positions: positions}
}

func typeKind(named *types.Named) domain.TypeKind {
	switch named.Underlying().(type) {
	case *types.Struct:
		return domain.KindStruct
	case *types.Interface:
		return domain.KindInterface
	default:
		return domain.KindOther
	}
}

func walkBody(req *walkBodyRequest) {
	if req.fieldCount > zero {
		req.facts.FieldsUsed = bitset.NewFieldSet(req.fieldCount)
	}

	if req.decl.Body == nil {
		return
	}

	selfIdx, hasSelf := req.methodIndex[req.self]
	ctx := siblingContext{siblings: make(map[int]bool), selfIdx: selfIdx, excludeSelf: hasSelf}
	ast.Inspect(req.decl.Body, walkBodyVisitor(req, ctx))
	assignSiblings(req.facts, ctx.siblings)
}

func walkBodyVisitor(req *walkBodyRequest, ctx siblingContext) func(ast.Node) bool {
	return func(n ast.Node) bool {
		countBranch(n, &req.facts.Branches)

		if sel, ok := n.(*ast.SelectorExpr); ok {
			recordSelection(&selectionRequest{
				info: req.info, n: sel, fieldIndex: req.fieldIndex,
				methodIndex: req.methodIndex, facts: req.facts, siblings: ctx.siblings,
				selfIdx: ctx.selfIdx, excludeSelf: ctx.excludeSelf,
			})
		}

		return true
	}
}

func assignSiblings(facts *domain.MethodFacts, siblings map[int]bool) {
	if len(siblings) == zero {
		return
	}

	facts.CalledSiblings = make([]int, zero, len(siblings))

	for idx := range siblings {
		facts.CalledSiblings = append(facts.CalledSiblings, idx)
	}

	slices.Sort(facts.CalledSiblings)
}
