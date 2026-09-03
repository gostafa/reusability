// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

func mergeTests(input *mergeTestsInput) ([]byte, error) {
	sortTestFiles(input.testFiles)

	state := newTestMergeState(input)

	err := collectAllTests(state, input)
	if err != nil {
		return nil, fmt.Errorf("collect tests: %w", err)
	}

	formatted, err := formatMergedTests(input, state)
	if err != nil {
		return nil, fmt.Errorf(fmtFormatMergedTest, err)
	}

	return formatted, nil
}

func formatMergedTests(input *mergeTestsInput, state *testMergeState) ([]byte, error) {
	file := &ast.File{
		Name:    ast.NewIdent(input.pkg.Name),
		Decls:   assembleTestDecls(state),
		Imports: collectImportSpecs(state.importSpecs),
	}

	formatted, err := formatWithImports(&formatInput{
		fset:    input.fset,
		license: input.license,
		file:    file,
		label:   "imports test",
	})
	if err != nil {
		return nil, fmt.Errorf(fmtFormatMergedTest, err)
	}

	return formatted, nil
}

func newTestMergeState(input *mergeTestsInput) *testMergeState {
	return &testMergeState{
		seenFuncs:   map[string]bool{},
		seenConsts:  map[string]string{},
		seenImports: map[string]bool{},
		external:    input.pkg.Name + "_test",
		dir:         input.pkg.Dir,
		pkgName:     input.pkg.Name,
		importPath:  input.pkg.ImportPath,
	}
}

func ingestOneDecl(state *testMergeState, req *testDeclIngest) {
	if tryIngestImportOrConst(state, req.decl, req.name) {
		return
	}

	appendKeptDecl(state, req)
}

func appendKeptDecl(state *testMergeState, req *testDeclIngest) {
	funcDecl, ok := req.decl.(*ast.FuncDecl)

	if !ok {
		state.otherDecls = append(state.otherDecls, req.decl)

		return
	}

	appendResolvedFunc(state, req, funcDecl)
}

func appendResolvedFunc(state *testMergeState, req *testDeclIngest, funcDecl *ast.FuncDecl) {
	kept := resolveDupFunc(state, &dupFuncIngest{
		funcDecl: funcDecl,
		name:     req.name,
		mode:     req.mode,
	})

	if kept == nil {
		return
	}

	state.otherDecls = append(state.otherDecls, kept)
}

func ingestTestDecls(state *testMergeState, req *testFileIngest) {
	for i := range req.file.Decls {
		ingestOneDecl(state, &testDeclIngest{
			decl: req.file.Decls[i],
			name: req.name,
			mode: req.mode,
		})
	}
}

func renameOrDropDup(state *testMergeState, req *dupFuncIngest) *ast.FuncDecl {
	if !allowExtRename(req.name, req.mode) {
		return nil
	}

	key := funcDedupKey(req.funcDecl)

	req.funcDecl.Name = ast.NewIdent(key + extNameSuffix)
	state.seenFuncs[req.funcDecl.Name.Name] = true

	return req.funcDecl
}

func resolveDupFunc(state *testMergeState, req *dupFuncIngest) *ast.FuncDecl {
	key := funcDedupKey(req.funcDecl)

	if !state.seenFuncs[key] {
		state.seenFuncs[key] = true

		return req.funcDecl
	}

	return renameOrDropDup(state, req)
}

func collectAllTests(state *testMergeState, input *mergeTestsInput) error {
	for i := range input.testFiles {
		err := collectOneTest(state, input.fset, input.testFiles[i])
		if err != nil {
			return fmt.Errorf("collect one test: %w", err)
		}
	}

	return nil
}

func collectOneTest(state *testMergeState, fset *token.FileSet, name string) error {
	path := filepath.Join(state.dir, name)

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse test %s: %w", name, err)
	}

	mode := prepareExternalTest(state, file)
	ingestTestDecls(state, &testFileIngest{file: file, name: name, mode: mode})

	return nil
}

func prepareExternalTest(state *testMergeState, file *ast.File) testFileMode {
	if file.Name.Name != state.external {
		return testFileInternal
	}

	aliases := stripSelfImport(file, state.importPath)

	for i := range aliases {
		rewriteExternalSelectors(file, aliases[i])
	}

	file.Name = ast.NewIdent(state.pkgName)

	return testFileExternal
}

func tryIngestImportOrConst(state *testMergeState, decl ast.Decl, name string) bool {
	if tryIngestImport(state, decl) {
		return true
	}

	return tryIngestConst(state, decl, name)
}

func tryIngestImport(state *testMergeState, decl ast.Decl) bool {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok || genDecl.Tok != token.IMPORT {
		return false
	}

	for i := range genDecl.Specs {
		ingestImportSpec(state, genDecl.Specs[i])
	}

	return true
}

func ingestImportSpec(state *testMergeState, spec ast.Spec) {
	importSpec, ok := spec.(*ast.ImportSpec)

	if !ok {
		return
	}

	key := importSpecKey(importSpec)

	if state.seenImports[key] {
		return
	}

	state.seenImports[key] = true
	state.importSpecs = append(state.importSpecs, importSpec)
}

func tryIngestConst(state *testMergeState, decl ast.Decl, name string) bool {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok || genDecl.Tok != token.CONST {
		return false
	}

	merged := mergeConstDecl(mergeConstInput{
		genDecl:    genDecl,
		seen:       state.seenConsts,
		onConflict: conflictActionFor(name),
	})

	if merged != nil {
		state.constDecls = append(state.constDecls, merged)
	}

	return true
}

func conflictActionFor(name string) constConflictAction {
	if strings.Contains(name, extTestSuffix) {
		return constConflictRename
	}

	return constConflictSkip
}

func allowExtRename(name string, mode testFileMode) bool {
	return mode == testFileExternal || strings.Contains(name, extTestSuffix)
}

func assembleTestDecls(state *testMergeState) []ast.Decl {
	var all []ast.Decl

	all = appendImportDecl(all, state.importSpecs)
	all = appendConstGroup(all, state.constDecls)
	all = append(all, state.otherDecls...)

	return all
}

func appendImportDecl(all []ast.Decl, specs []ast.Spec) []ast.Decl {
	if len(specs) == countZero {
		return all
	}

	return append(all, &ast.GenDecl{Tok: token.IMPORT, Specs: specs})
}

func appendConstGroup(all, constDecls []ast.Decl) []ast.Decl {
	specs := flattenConstSpecs(constDecls)

	if len(specs) == countZero {
		return all
	}

	return append(all, groupedGenDecl(token.CONST, specs))
}

func flattenConstSpecs(constDecls []ast.Decl) []ast.Spec {
	var specs []ast.Spec

	for i := range constDecls {
		specs = appendConstSpecs(specs, constDecls[i])
	}

	return specs
}

func appendConstSpecs(specs []ast.Spec, decl ast.Decl) []ast.Spec {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok {
		return specs
	}

	return append(specs, genDecl.Specs...)
}

func sortTestFiles(files []string) {
	slices.SortStableFunc(files, compareTestFiles)
}

func compareTestFiles(left, right string) int {
	leftPrio, rightPrio := testFilePriority(left), testFilePriority(right)

	if leftPrio != rightPrio {
		return cmp.Compare(leftPrio, rightPrio)
	}

	return cmp.Compare(left, right)
}

func testFilePriority(name string) int {
	switch {
	case strings.HasPrefix(name, constantsTest):
		return countZero
	case strings.HasSuffix(name, extTestSuffix):
		return countThree
	case strings.Contains(name, benchFragment):
		return countTwo
	default:
		return countOne
	}
}

func funcDedupKey(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil {
		return methodName(funcDecl)
	}

	return funcDecl.Name.Name
}
