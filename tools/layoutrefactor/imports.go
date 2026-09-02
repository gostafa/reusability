// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"cmp"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

func collectProdImports(fset *token.FileSet, dir string, prodFiles []string) []ast.Spec {
	byPath := map[string]*ast.ImportSpec{}

	for i := range prodFiles {
		mergeFileImports(&mergeFileImportsInput{
			byPath: byPath,
			fset:   fset,
			dir:    dir,
			name:   prodFiles[i],
		})
	}

	return sortedImportSpecs(byPath)
}

func mergeFileImports(input *mergeFileImportsInput) {
	path := filepath.Join(input.dir, input.name)

	file, err := parser.ParseFile(input.fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return
	}

	for i := range file.Imports {
		storeImport(input.byPath, file.Imports[i])
	}
}

func storeImport(byPath map[string]*ast.ImportSpec, imp *ast.ImportSpec) {
	path := strings.Trim(imp.Path.Value, quoteChar)

	if _, ok := byPath[path]; ok {
		return
	}

	byPath[path] = imp
}

func sortedImportSpecs(byPath map[string]*ast.ImportSpec) []ast.Spec {
	specs := make([]ast.Spec, countZero, len(byPath))

	for path := range byPath {
		specs = append(specs, byPath[path])
	}

	slices.SortFunc(specs, compareImportSpecs)

	return specs
}

func compareImportSpecs(left, right ast.Spec) int {
	leftImp, leftOK := left.(*ast.ImportSpec)
	rightImp, rightOK := right.(*ast.ImportSpec)

	if !leftOK || !rightOK {
		return countZero
	}

	return cmp.Compare(importSpecKey(leftImp), importSpecKey(rightImp))
}

func importSpecKey(importSpec *ast.ImportSpec) string {
	path := strings.Trim(importSpec.Path.Value, quoteChar)

	if importSpec.Name != nil {
		return importSpec.Name.Name + " " + path
	}

	return path
}

func collectImportSpecs(specs []ast.Spec) []*ast.ImportSpec {
	var out []*ast.ImportSpec

	for i := range specs {
		out = appendImportSpec(out, specs[i])
	}

	return out
}

func appendImportSpec(out []*ast.ImportSpec, spec ast.Spec) []*ast.ImportSpec {
	importSpec, ok := spec.(*ast.ImportSpec)

	if !ok {
		return out
	}

	return append(out, importSpec)
}

func stripSelfImport(file *ast.File, importPath string) []string {
	if len(file.Imports) == countZero {
		return nil
	}

	aliases, kept := filterSelfImports(file.Imports, importPath)

	file.Imports = kept
	file.Decls = removeImportDecls(file.Decls, importPath)

	return aliases
}

func filterSelfImports(imports []*ast.ImportSpec, importPath string) ([]string, []*ast.ImportSpec) {
	return collectFilteredImports(&filterImportsInput{
		imports:    imports,
		importPath: importPath,
	})
}

func collectFilteredImports(input *filterImportsInput) ([]string, []*ast.ImportSpec) {
	var (
		aliases []string
		kept    []*ast.ImportSpec
	)

	for i := range input.imports {
		next := filterOneImport(&filterImportInput{
			aliases:    aliases,
			kept:       kept,
			imp:        input.imports[i],
			importPath: input.importPath,
		})

		aliases, kept = next.aliases, next.kept
	}

	return aliases, kept
}

func filterOneImport(input *filterImportInput) filterImportInput {
	path := strings.Trim(input.imp.Path.Value, quoteChar)

	if path != input.importPath {
		input.kept = append(input.kept, input.imp)

		return *input
	}

	input.aliases = append(input.aliases, importAlias(input.imp, input.importPath))

	return *input
}

func importAlias(imp *ast.ImportSpec, importPath string) string {
	if imp.Name != nil {
		return imp.Name.Name
	}

	return filepath.Base(importPath)
}

func removeImportDecls(decls []ast.Decl, importPath string) []ast.Decl {
	var out []ast.Decl

	for i := range decls {
		out = appendFilteredDecl(out, decls[i], importPath)
	}

	return out
}

func appendFilteredDecl(out []ast.Decl, decl ast.Decl, importPath string) []ast.Decl {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok || genDecl.Tok != token.IMPORT {
		return append(out, decl)
	}

	return appendFilteredImport(out, genDecl, importPath)
}

func appendFilteredImport(out []ast.Decl, genDecl *ast.GenDecl, importPath string) []ast.Decl {
	specs := filterImportSpecs(genDecl.Specs, importPath)

	if len(specs) == countZero {
		return out
	}

	genDecl.Specs = specs

	return append(out, genDecl)
}

func filterImportSpecs(specs []ast.Spec, importPath string) []ast.Spec {
	var out []ast.Spec

	for i := range specs {
		out = keepImportSpec(out, specs[i], importPath)
	}

	return out
}

func keepImportSpec(out []ast.Spec, spec ast.Spec, importPath string) []ast.Spec {
	importSpec, ok := spec.(*ast.ImportSpec)

	if !ok {
		return append(out, spec)
	}

	if strings.Trim(importSpec.Path.Value, quoteChar) == importPath {
		return out
	}

	return append(out, spec)
}
