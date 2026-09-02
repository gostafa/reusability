// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"go/ast"
	"reflect"
)

func rewriteExternalSelectors(file *ast.File, pkgIdent string) {
	ast.Inspect(file, func(node ast.Node) bool {
		return rewriteSelectorNode(file, node, pkgIdent)
	})
}

func rewriteSelectorNode(file *ast.File, node ast.Node, pkgIdent string) bool {
	sel, ok := node.(*ast.SelectorExpr)

	if !ok {
		return true
	}

	ident, ok := sel.X.(*ast.Ident)

	if !ok || ident.Name != pkgIdent {
		return true
	}

	replacement := ast.NewIdent(sel.Sel.Name)

	if !replaceChild(file, sel, replacement) {
		sel.X = replacement
	}

	return true
}

func replaceChild(root, old, replacement ast.Node) bool {
	replaced := false
	input := &replaceInput{old: old, replacement: replacement, replaced: &replaced}

	ast.Inspect(root, func(node ast.Node) bool {
		return inspectReplace(node, input)
	})

	return replaced
}

func inspectReplace(node ast.Node, input *replaceInput) bool {
	if *input.replaced || node == nil {
		return !*input.replaced
	}

	return scanNodeFields(node, input)
}

func scanNodeFields(node ast.Node, input *replaceInput) bool {
	value := reflect.ValueOf(node)

	if value.Kind() != reflect.Pointer || value.IsNil() {
		return true
	}

	value = value.Elem()

	if !value.IsValid() {
		return true
	}

	return scanStructFields(value, input)
}

func scanStructFields(value reflect.Value, input *replaceInput) bool {
	if replaceFieldAt(value, input, value.NumField()) {
		*input.replaced = true

		return false
	}

	return true
}

func replaceFieldAt(value reflect.Value, input *replaceInput, remaining int) bool {
	if remaining == countZero {
		return false
	}

	idx := remaining - countOne

	if tryReplaceField(value.Field(idx), input) {
		return true
	}

	return replaceFieldAt(value, input, idx)
}

func tryReplaceField(field reflect.Value, input *replaceInput) bool {
	if !field.CanSet() {
		return false
	}

	if replacePointerField(field, input) {
		return true
	}

	return replaceSliceField(field, input)
}

func replacePointerField(field reflect.Value, input *replaceInput) bool {
	if field.Kind() != reflect.Interface && field.Kind() != reflect.Pointer {
		return false
	}

	if field.IsNil() || field.Interface() != input.old {
		return false
	}

	field.Set(reflect.ValueOf(input.replacement))

	return true
}

func replaceSliceField(field reflect.Value, input *replaceInput) bool {
	if field.Kind() != reflect.Slice {
		return false
	}

	for idx := range field.Len() {
		if replaceSliceElem(field.Index(idx), input) {
			return true
		}
	}

	return false
}

func replaceSliceElem(elem reflect.Value, input *replaceInput) bool {
	if !elem.CanInterface() || elem.Interface() != input.old {
		return false
	}

	elem.Set(reflect.ValueOf(input.replacement))

	return true
}
