// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func refactorPackage(pkg *packageInfo, opts runOptions) error {
	scan, err := scanPackage(pkg)
	if err != nil {
		return fmt.Errorf("scan package: %w", err)
	}

	if len(scan.prodFiles) == countZero && len(scan.testFiles) == countZero {
		return nil
	}

	err = applyRefactor(pkg, scan, opts)
	if err != nil {
		return fmt.Errorf("apply refactor: %w", err)
	}

	return nil
}

func scanPackage(pkg *packageInfo) (*prodScanResult, error) {
	entries, err := os.ReadDir(pkg.Dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	prod, test := splitGoFiles(entries)

	result, err := buildScanResult(pkg, prod, test)
	if err != nil {
		return nil, fmt.Errorf("build scan: %w", err)
	}

	return result, nil
}

func splitGoFiles(entries []os.DirEntry) (prodFiles, testFiles []string) {
	for i := range entries {
		prodFiles, testFiles = classifyGoEntry(entries[i], prodFiles, testFiles)
	}

	return prodFiles, testFiles
}

func classifyGoEntry(entry os.DirEntry, prod, test []string) (prodFiles, testFiles []string) {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), goSuffix) {
		return prod, test
	}

	if strings.HasSuffix(entry.Name(), testSuffix) {
		return prod, append(test, entry.Name())
	}

	return append(prod, entry.Name()), test
}

func buildScanResult(pkg *packageInfo, prod, test []string) (*prodScanResult, error) {
	fset := token.NewFileSet()

	decls, doc, err := parseProdDecls(&prodParseInput{
		fset: fset,
		dir:  pkg.Dir,
		prod: prod,
	})
	if err != nil {
		return nil, fmt.Errorf("parse prod decls: %w", err)
	}

	return newProdScanResult(&prodScanBuild{
		pkg:   pkg,
		fset:  fset,
		doc:   doc,
		prod:  prod,
		test:  test,
		decls: decls,
	}), nil
}

func newProdScanResult(build *prodScanBuild) *prodScanResult {
	return &prodScanResult{
		prodFiles:  build.prod,
		testFiles:  build.test,
		allDecls:   build.decls,
		packageDoc: choosePackageDoc(build.doc, build.pkg.Name),
		pkgImports: collectProdImports(build.fset, build.pkg.Dir, build.prod),
		license:    extractLicense(),
		fset:       build.fset,
		pkgName:    build.pkg.Name,
	}
}

func parseProdDecls(input *prodParseInput) (decls []pkgDecl, doc string, err error) {
	allDecls, packageDoc, err := accumulateProdDecls(input)
	if err != nil {
		return nil, emptyString, fmt.Errorf("accumulate prod: %w", err)
	}

	return allDecls, packageDoc, nil
}

func accumulateProdDecls(input *prodParseInput) (decls []pkgDecl, doc string, err error) {
	state := &prodAccumulateState{}

	for i := range input.prod {
		err = accumulateOneProd(input, state, input.prod[i])
		if err != nil {
			return nil, emptyString, fmt.Errorf("accumulate one: %w", err)
		}
	}

	return state.decls, state.doc, nil
}

func accumulateOneProd(input *prodParseInput, state *prodAccumulateState, name string) error {
	next, doc, err := parseAndMergeProd(&prodFileInput{
		fset:       input.fset,
		dir:        input.dir,
		name:       name,
		packageDoc: state.doc,
	})
	if err != nil {
		return fmt.Errorf("parse and merge: %w", err)
	}

	state.doc = doc
	state.decls = append(state.decls, next...)

	return nil
}

func parseAndMergeProd(input *prodFileInput) (decls []pkgDecl, doc string, err error) {
	decls, doc, err = parseOneProdFile(input)
	if err != nil {
		return nil, emptyString, fmt.Errorf("parse prod file: %w", err)
	}

	return decls, mergePackageDoc(input.packageDoc, doc, input.name), nil
}

func parseOneProdFile(input *prodFileInput) (decls []pkgDecl, doc string, err error) {
	path := filepath.Join(input.dir, input.name)

	file, err := parser.ParseFile(input.fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, emptyString, fmt.Errorf("parse %s: %w", input.name, err)
	}

	return collectFileDecls(file), formatPackageDoc(file.Doc), nil
}

func collectFileDecls(file *ast.File) []pkgDecl {
	var out []pkgDecl

	for i := range file.Decls {
		out = appendDecl(out, file.Decls[i])
	}

	return out
}

func appendDecl(out []pkgDecl, decl ast.Decl) []pkgDecl {
	kind, skip := classifyDecl(decl)

	if skip {
		return out
	}

	return append(out, pkgDecl{kind: kind, decl: decl})
}

func mergePackageDoc(current, doc, name string) string {
	if doc == emptyString {
		return current
	}

	if current == emptyString || name == docGoName {
		return doc
	}

	return current
}

func choosePackageDoc(doc, pkgName string) string {
	if doc == emptyString {
		return defaultPackageDoc(pkgName)
	}

	return doc
}

func applyRefactor(pkg *packageInfo, scan *prodScanResult, opts runOptions) error {
	writes, err := buildAllWrites(pkg, scan)
	if err != nil {
		return fmt.Errorf("build writes: %w", err)
	}

	deletes := planDeletes(scan.prodFiles, scan.testFiles, testFileName(pkg.Name))

	err = commitWrites(&commitInput{pkg: pkg, writes: writes, deletes: deletes, opts: opts})
	if err != nil {
		return fmt.Errorf(fmtCommitWrites, err)
	}

	return nil
}

func buildAllWrites(pkg *packageInfo, scan *prodScanResult) ([]writeOp, error) {
	writes, err := buildDocAndProdWrites(pkg, scan)
	if err != nil {
		return nil, fmt.Errorf("doc and prod writes: %w", err)
	}

	result, err := appendTestWrite(writes, pkg, scan)
	if err != nil {
		return nil, fmt.Errorf("append test write: %w", err)
	}

	return result, nil
}

func buildDocAndProdWrites(pkg *packageInfo, scan *prodScanResult) ([]writeOp, error) {
	prodWrites, err := buildProdWrites(scan)
	if err != nil {
		return nil, fmt.Errorf("build prod writes: %w", err)
	}

	docContent, err := buildDocGo(scan.license, pkg.Name, scan.packageDoc)
	if err != nil {
		return nil, fmt.Errorf("build doc.go: %w", err)
	}

	writes := make([]writeOp, countZero, countOne+len(prodWrites))

	writes = append(writes, writeOp{name: docGoName, content: docContent})

	return append(writes, prodWrites...), nil
}

func buildProdWrites(scan *prodScanResult) ([]writeOp, error) {
	byKind := groupDeclsByKind(scan.allDecls)
	targets := prodTargets()
	kinds := kindOrder()

	var writes []writeOp

	for i := range kinds {
		kindOp, err := buildKindWrite(&kindWriteInput{
			scan:  scan,
			decls: byKind[kinds[i]],
			name:  targets[kinds[i]],
		})
		if err != nil {
			return nil, fmt.Errorf("build kind write: %w", err)
		}

		writes = appendOptionalWrite(writes, &kindOp)
	}

	return writes, nil
}

func kindOrder() []declKind {
	return []declKind{kindConst, kindType, kindVar, kindFunc}
}

func appendOptionalWrite(writes []writeOp, op *writeOp) []writeOp {
	if op.name == emptyString {
		return writes
	}

	return append(writes, *op)
}

func groupDeclsByKind(all []pkgDecl) map[declKind][]ast.Decl {
	byKind := map[declKind][]ast.Decl{}

	for i := range all {
		byKind[all[i].kind] = append(byKind[all[i].kind], all[i].decl)
	}

	return byKind
}

func prodTargets() map[declKind]string {
	return map[declKind]string{
		kindConst: constsGoName,
		kindType:  typesGoName,
		kindVar:   varsGoName,
		kindFunc:  funcsGoName,
	}
}

func buildKindWrite(input *kindWriteInput) (writeOp, error) {
	if len(input.decls) == countZero {
		return writeOp{}, nil
	}

	op, err := formatKindWrite(input)
	if err != nil {
		return writeOp{}, fmt.Errorf("format kind write: %w", err)
	}

	return op, nil
}

func formatKindWrite(input *kindWriteInput) (writeOp, error) {
	content, err := buildDeclFile(&declFileInput{
		fset:       input.scan.fset,
		license:    input.scan.license,
		pkgName:    input.scan.pkgName,
		decls:      input.decls,
		pkgImports: input.scan.pkgImports,
	})
	if err != nil {
		return writeOp{}, fmt.Errorf("build decl file: %w", err)
	}

	return writeOp{name: input.name, content: content}, nil
}

func appendTestWrite(writes []writeOp, pkg *packageInfo, scan *prodScanResult) ([]writeOp, error) {
	if len(scan.testFiles) == countZero {
		return writes, nil
	}

	merged, err := mergeTests(&mergeTestsInput{
		fset:      scan.fset,
		pkg:       *pkg,
		testFiles: scan.testFiles,
		license:   scan.license,
	})
	if err != nil {
		return nil, fmt.Errorf("merge tests: %w", err)
	}

	return append(writes, writeOp{name: testFileName(pkg.Name), content: merged}), nil
}

func testFileName(pkgName string) string {
	if pkgName == mainPkgName {
		return mainTestName
	}

	return pkgName + testSuffix
}
