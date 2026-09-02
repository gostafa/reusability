// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"

	"golang.org/x/tools/imports"
)

func consolidateGenDecls(decls []ast.Decl) []ast.Decl {
	buckets := &genDeclBuckets{}
	result := []ast.Decl{}

	for i := range decls {
		result = accumulateGenDecl(result, buckets, decls[i])
	}

	return appendGroupedDecls(result, buckets)
}

func accumulateGenDecl(out []ast.Decl, buckets *genDeclBuckets, decl ast.Decl) []ast.Decl {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok {
		return append(out, decl)
	}

	return bucketGenDecl(out, buckets, genDecl)
}

func bucketGenDecl(input []ast.Decl, buckets *genDeclBuckets, genDecl *ast.GenDecl) []ast.Decl {
	if genDecl.Tok == token.CONST {
		buckets.constSpecs = append(buckets.constSpecs, genDecl.Specs...)

		return input
	}

	if genDecl.Tok == token.TYPE {
		buckets.typeSpecs = append(buckets.typeSpecs, genDecl.Specs...)

		return input
	}

	if genDecl.Tok == token.VAR {
		buckets.varSpecs = append(buckets.varSpecs, genDecl.Specs...)

		return input
	}

	return append(input, genDecl)
}

func appendGroupedDecls(input []ast.Decl, buckets *genDeclBuckets) []ast.Decl {
	result := appendIfGrouped(input, token.TYPE, buckets.typeSpecs)

	result = appendIfGrouped(result, token.CONST, buckets.constSpecs)

	return appendIfGrouped(result, token.VAR, buckets.varSpecs)
}

func appendIfGrouped(out []ast.Decl, tok token.Token, specs []ast.Spec) []ast.Decl {
	if len(specs) == countZero {
		return out
	}

	return append(out, groupedGenDecl(tok, specs))
}

func groupedGenDecl(tok token.Token, specs []ast.Spec) *ast.GenDecl {
	if len(specs) == countOne {
		return &ast.GenDecl{Tok: tok, Specs: specs}
	}

	return &ast.GenDecl{Tok: tok, Lparen: token.Pos(countOne), Specs: specs}
}

func buildDeclFile(input *declFileInput) ([]byte, error) {
	decls := consolidateGenDecls(input.decls)
	sortDecls(decls)

	file := &ast.File{
		Name:  ast.NewIdent(input.pkgName),
		Decls: withImportDecls(decls, input.pkgImports),
	}

	formatted, err := formatWithImports(&formatInput{
		fset:    input.fset,
		license: input.license,
		file:    file,
		label:   "imports",
	})
	if err != nil {
		return nil, fmt.Errorf("format decl file: %w", err)
	}

	return formatted, nil
}

func withImportDecls(decls []ast.Decl, pkgImports []ast.Spec) []ast.Decl {
	if len(pkgImports) == countZero {
		return decls
	}

	imp := &ast.GenDecl{Tok: token.IMPORT, Specs: pkgImports}

	return append([]ast.Decl{imp}, decls...)
}

func formatWithImports(input *formatInput) ([]byte, error) {
	var raw bytes.Buffer

	err := format.Node(&raw, input.fset, input.file)
	if err != nil {
		return nil, fmt.Errorf("format node: %w", err)
	}

	src := input.license + doubleNewline + raw.String()

	formatted, err := imports.Process(emptyString, []byte(src), &imports.Options{Comments: true})
	if err != nil {
		return nil, fmt.Errorf("%s: %w\nsource:\n%s", input.label, err, src)
	}

	return formatted, nil
}
