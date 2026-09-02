// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"cmp"
	"go/ast"
	"slices"
	"strings"
	"unicode"
)

func declName(decl ast.Decl) string {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return funcDeclName(typed)
	case *ast.GenDecl:
		return genDeclName(typed)
	default:
		return emptyString
	}
}

func funcDeclName(decl *ast.FuncDecl) string {
	if decl.Recv != nil {
		return methodName(decl)
	}

	return decl.Name.Name
}

func genDeclName(decl *ast.GenDecl) string {
	for i := range decl.Specs {
		if name := specName(decl.Specs[i]); name != emptyString {
			return name
		}
	}

	return emptyString
}

func specName(spec ast.Spec) string {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		return typed.Name.Name
	case *ast.ValueSpec:
		return valueSpecName(typed)
	default:
		return emptyString
	}
}

func valueSpecName(spec *ast.ValueSpec) string {
	if len(spec.Names) == countZero {
		return emptyString
	}

	return spec.Names[countZero].Name
}

func methodName(funcDecl *ast.FuncDecl) string {
	recv := recvTypeName(funcDecl)

	if recv == emptyString {
		return funcDecl.Name.Name
	}

	return recv + dot + funcDecl.Name.Name
}

func recvTypeName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == countZero {
		return emptyString
	}

	return typeIdentName(funcDecl.Recv.List[countZero].Type)
}

func typeIdentName(typ ast.Expr) string {
	switch typed := typ.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return starIdentName(typed)
	default:
		return emptyString
	}
}

func starIdentName(star *ast.StarExpr) string {
	ident, ok := star.X.(*ast.Ident)

	if !ok {
		return emptyString
	}

	return ident.Name
}

func isExported(name string) bool {
	if name == emptyString {
		return false
	}

	if strings.Contains(name, dot) {
		return exportedSelector(name)
	}

	return unicode.IsUpper(rune(name[countZero]))
}

func exportedSelector(name string) bool {
	parts := strings.SplitN(name, dot, countTwo)

	return unicode.IsUpper(rune(parts[countZero][countZero])) ||
		unicode.IsUpper(rune(parts[countOne][countZero]))
}

func sortDecls(decls []ast.Decl) {
	slices.SortStableFunc(decls, compareDecls)
}

func compareDecls(left, right ast.Decl) int {
	leftName, rightName := declName(left), declName(right)
	leftExp, rightExp := isExported(leftName), isExported(rightName)

	if leftExp == rightExp {
		return cmp.Compare(leftName, rightName)
	}

	if leftExp {
		return cmpLess
	}

	return countOne
}
