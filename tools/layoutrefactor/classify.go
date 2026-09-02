// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"go/ast"
	"go/token"
)

func classifyDecl(decl ast.Decl) (declKind, bool) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return kindFunc, false
	case *ast.GenDecl:
		return classifyGenDecl(typed)
	default:
		return kindFunc, true
	}
}

func classifyGenDecl(decl *ast.GenDecl) (declKind, bool) {
	if decl.Tok == token.CONST {
		return kindConst, false
	}

	if decl.Tok == token.TYPE {
		return kindType, false
	}

	if decl.Tok == token.VAR {
		return kindVar, false
	}

	return kindFunc, true
}
