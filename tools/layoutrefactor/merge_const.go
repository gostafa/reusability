// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

func mergeConstDecl(input mergeConstInput) *ast.GenDecl {
	var kept []ast.Spec

	for i := range input.genDecl.Specs {
		kept = keepConstSpec(kept, input.genDecl.Specs[i], input)
	}

	if len(kept) == countZero {
		return nil
	}

	return &ast.GenDecl{Tok: token.CONST, Lparen: input.genDecl.Lparen, Specs: kept}
}

func keepConstSpec(kept []ast.Spec, spec ast.Spec, input mergeConstInput) []ast.Spec {
	valueSpec, ok := spec.(*ast.ValueSpec)

	if !ok || len(valueSpec.Names) == countZero {
		return append(kept, spec)
	}

	return keepNamedConst(kept, valueSpec, input)
}

func keepNamedConst(kept []ast.Spec, valueSpec *ast.ValueSpec, input mergeConstInput) []ast.Spec {
	name := valueSpec.Names[countZero].Name
	val := constSpecLiteral(valueSpec)

	if prev, ok := input.seen[name]; ok {
		return resolveConstConflict(&constConflictInput{
			kept:      kept,
			valueSpec: valueSpec,
			input:     input,
			name:      name,
			val:       val,
			prev:      prev,
		})
	}

	input.seen[name] = val

	return append(kept, valueSpec)
}

func resolveConstConflict(conflict *constConflictInput) []ast.Spec {
	if conflict.prev == conflict.val {
		return conflict.kept
	}

	if conflict.input.onConflict != constConflictRename {
		return conflict.kept
	}

	renamed := renameExtConst(conflict.valueSpec, conflict.name)

	conflict.input.seen[renamed] = conflict.val

	return append(conflict.kept, conflict.valueSpec)
}

func renameExtConst(valueSpec *ast.ValueSpec, name string) string {
	renamed := extPrefix + strings.ToUpper(name[:countOne]) + name[countOne:]

	valueSpec.Names[countZero] = ast.NewIdent(renamed)

	return renamed
}

func constSpecLiteral(valueSpec *ast.ValueSpec) string {
	if len(valueSpec.Values) == countZero {
		return emptyString
	}

	var buf bytes.Buffer

	err := format.Node(&buf, token.NewFileSet(), valueSpec.Values[countZero])
	if err != nil {
		return emptyString
	}

	return buf.String()
}

func uniqueStrings(input []string) []string {
	seen := map[string]bool{}

	var out []string

	for i := range input {
		out = appendUnique(out, seen, input[i])
	}

	return out
}

func appendUnique(out []string, seen map[string]bool, item string) []string {
	if seen[item] {
		return out
	}

	seen[item] = true

	return append(out, item)
}
