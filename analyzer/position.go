package analyzer

import (
	"go/ast"
	"go/token"

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
