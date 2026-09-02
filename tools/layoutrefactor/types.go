// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"go/ast"
	"go/token"
)

type (
	declKind int

	pkgDecl struct {
		decl ast.Decl
		kind declKind
	}

	packageInfo struct {
		ImportPath string
		Dir        string
		Name       string
	}

	writeOp struct {
		name    string
		content []byte
	}

	constConflictAction int

	declFileInput struct {
		fset       *token.FileSet
		license    string
		pkgName    string
		decls      []ast.Decl
		pkgImports []ast.Spec
	}

	mergeTestsInput struct {
		fset      *token.FileSet
		pkg       packageInfo
		license   string
		testFiles []string
	}

	mergeConstInput struct {
		genDecl    *ast.GenDecl
		seen       map[string]string
		onConflict constConflictAction
	}

	testMergeState struct {
		seenFuncs   map[string]bool
		seenConsts  map[string]string
		seenImports map[string]bool
		external    string
		dir         string
		pkgName     string
		importPath  string
		importSpecs []ast.Spec
		constDecls  []ast.Decl
		otherDecls  []ast.Decl
	}

	prodScanResult struct {
		fset       *token.FileSet
		packageDoc string
		license    string
		pkgName    string
		prodFiles  []string
		testFiles  []string
		allDecls   []pkgDecl
		pkgImports []ast.Spec
	}

	replaceInput struct {
		old         ast.Node
		replacement ast.Node
		replaced    *bool
	}

	formatInput struct {
		fset    *token.FileSet
		license string
		file    *ast.File
		label   string
	}

	mergeFileImportsInput struct {
		byPath map[string]*ast.ImportSpec
		fset   *token.FileSet
		dir    string
		name   string
	}

	filterImportInput struct {
		imp        *ast.ImportSpec
		importPath string
		aliases    []string
		kept       []*ast.ImportSpec
	}

	ingestDeclInput struct {
		state *testMergeState
		decl  ast.Decl
		name  string
		mode  testFileMode
	}

	testFileMode int

	constConflictInput struct {
		valueSpec *ast.ValueSpec
		input     mergeConstInput
		name      string
		val       string
		prev      string
		kept      []ast.Spec
	}

	genDeclBuckets struct {
		constSpecs []ast.Spec
		typeSpecs  []ast.Spec
		varSpecs   []ast.Spec
	}

	runOptions struct {
		pattern string
		dryRun  bool
	}

	commitInput struct {
		pkg     *packageInfo
		writes  []writeOp
		deletes []string
		opts    runOptions
	}

	docSectionsInput struct {
		license string
		pkgName string
		doc     string
	}

	prodParseInput struct {
		fset *token.FileSet
		dir  string
		prod []string
	}

	prodFileInput struct {
		fset       *token.FileSet
		dir        string
		name       string
		packageDoc string
	}

	kindWriteInput struct {
		scan  *prodScanResult
		name  string
		decls []ast.Decl
	}

	testWriteInput struct {
		pkg    *packageInfo
		scan   *prodScanResult
		writes []writeOp
	}

	filterImportsInput struct {
		importPath string
		imports    []*ast.ImportSpec
	}

	scanAssembleInput struct {
		pkg   *packageInfo
		fset  *token.FileSet
		doc   string
		prod  []string
		test  []string
		decls []pkgDecl
	}

	prodAccumulateState struct {
		doc   string
		decls []pkgDecl
	}
)
