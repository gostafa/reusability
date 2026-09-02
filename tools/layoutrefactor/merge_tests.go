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
	ingestTestDecls(state, name, mode, file)

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

func ingestTestDecls(state *testMergeState, name string, mode testFileMode, file *ast.File) {
	for i := range file.Decls {
		ingestOneDecl(state, file.Decls[i], name, mode)
	}
}

func ingestOneDecl(state *testMergeState, decl ast.Decl, name string, mode testFileMode) {
	if tryIngestImportOrConst(state, decl, name) {
		return
	}

	funcDecl, ok := decl.(*ast.FuncDecl)

	if !ok {
		state.otherDecls = append(state.otherDecls, decl)

		return
	}

	kept := resolveDupFunc(state, funcDecl, name, mode)

	if kept == nil {
		return
	}

	state.otherDecls = append(state.otherDecls, kept)
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

func resolveDupFunc(
	state *testMergeState,
	funcDecl *ast.FuncDecl,
	name string,
	mode testFileMode,
) *ast.FuncDecl {
	key := funcDedupKey(funcDecl)

	if !state.seenFuncs[key] {
		state.seenFuncs[key] = true

		return funcDecl
	}

	return renameOrDropDup(state, funcDecl, key, name, mode)
}

func renameOrDropDup(
	state *testMergeState,
	funcDecl *ast.FuncDecl,
	key, name string,
	mode testFileMode,
) *ast.FuncDecl {
	if !allowExtRename(name, mode) {
		return nil
	}

	funcDecl.Name = ast.NewIdent(key + extNameSuffix)
	state.seenFuncs[funcDecl.Name.Name] = true

	return funcDecl
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
