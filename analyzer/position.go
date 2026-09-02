package analyzer

import (
	"go/ast"
	"go/token"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

// packagePos returns a position for package-scoped diagnostics: the package
// clause of the first file, or token.NoPos when the pass has no files.
func packagePos(pass *analysis.Pass) token.Pos {
	for _, file := range pass.Files {
		if file != nil {
			return file.Package
		}
	}

	return token.NoPos
}

// typePos returns the position of the named type's TypeSpec identifier, or
// the package position when the type is not found in the pass files.
func typePos(pass *analysis.Pass, name string) token.Pos {
	for _, file := range pass.Files {
		if file == nil {
			continue
		}

		if pos := findTypeInFile(file, name); pos != token.NoPos {
			return pos
		}
	}

	return packagePos(pass)
}

// structuralPos returns a position for package-level structural diagnostics
// whose key names a concrete declaration category.
func structuralPos(pass *analysis.Pass, key string) token.Pos {
	switch key {
	case policydomain.KeyFuncs:
		return funcPos(pass, nil)
	case policydomain.KeyExportedFuncs:
		exported := true

		return funcPos(pass, &exported)
	case policydomain.KeyUnexportedFuncs:
		exported := false

		return funcPos(pass, &exported)
	case policydomain.KeyVars:
		return valuePos(pass, token.VAR)
	case policydomain.KeyConsts:
		return valuePos(pass, token.CONST)
	default:
		return packagePos(pass)
	}
}

func funcPos(pass *analysis.Pass, exported *bool) token.Pos {
	for _, file := range pass.Files {
		if file == nil {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}

			if exported == nil || fn.Name.IsExported() == *exported {
				return fn.Name.Pos()
			}
		}
	}

	return packagePos(pass)
}

func exactFuncPos(pass *analysis.Pass, receiver, name string) token.Pos {
	for _, file := range pass.Files {
		if file == nil {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != name {
				continue
			}

			if receiverTypeName(fn.Recv) == receiver {
				return fn.Name.Pos()
			}
		}
	}

	if receiver != "" {
		return typePos(pass, receiver)
	}

	return packagePos(pass)
}

func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	return receiverExprName(recv.List[0].Type)
}

func receiverExprName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverExprName(typed.X)
	case *ast.IndexExpr:
		return receiverExprName(typed.X)
	case *ast.IndexListExpr:
		return receiverExprName(typed.X)
	default:
		return ""
	}
}

func valuePos(pass *analysis.Pass, tok token.Token) token.Pos {
	for _, file := range pass.Files {
		if file == nil {
			continue
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != tok {
				continue
			}

			if pos := firstValueNamePos(gen); pos != token.NoPos {
				return pos
			}
		}
	}

	return packagePos(pass)
}

func firstValueNamePos(decl *ast.GenDecl) token.Pos {
	for _, spec := range decl.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for _, name := range value.Names {
			if name != nil && name.Name != "_" {
				return name.Pos()
			}
		}
	}

	return token.NoPos
}

func findTypeInFile(file *ast.File, name string) token.Pos {
	var found token.Pos

	ast.Inspect(file, func(n ast.Node) bool {
		if found != token.NoPos {
			return false
		}

		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name == nil || spec.Name.Name != name {
			return true
		}

		found = spec.Name.Pos()

		return false
	})

	return found
}
