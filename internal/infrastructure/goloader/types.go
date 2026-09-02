// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/workerpool"
	"golang.org/x/tools/go/packages"
)

type (
	generatedInclusion bool
	packageErrorPolicy bool

	loaderDeps struct {
		packagesLoad      func(*packages.Config, ...string) ([]*packages.Package, error)
		runExtractWorkers func(context.Context, workerpool.RunConfig) error
		getwd             func() (string, error)
		absPath           func(string) (string, error)
	}

	extractorOptions struct {
		analyzed   map[string]bool
		modulePath string
		baseDir    string
		generated  generatedInclusion
	}

	docRange struct {
		start, end token.Pos
		documented bool
	}

	methodDecl struct {
		fn   *types.Func
		decl *ast.FuncDecl
	}

	methodDocInput struct {
		exported   bool
		documented bool
	}

	refCollector struct {
		self     *types.TypeName
		analyzed map[string]bool
		seen     map[string]bool
		visited  map[types.Type]bool
	}

	syntaxIndex struct {
		generated map[string]bool
		funcDecls map[*types.Func]*ast.FuncDecl
		typeDocs  map[types.Object]bool
		fieldDocs []docRange
	}

	extractJob struct {
		opts       *outbound.FactOptions
		modulePath string
		pkgs       []*packages.Package
	}

	loadedPackages struct {
		opts *outbound.FactOptions
		deps *loaderDeps
		pkgs []*packages.Package
	}

	selectLoadedRequest struct {
		loaded   []*packages.Package
		opts     *outbound.FactOptions
		patterns []string
	}

	filterPackagesRequest struct {
		byPath map[string]*packages.Package
		order  []string
		policy packageErrorPolicy
	}

	structFieldsResult struct {
		fields     []domain.FieldFacts
		fieldIndex map[*types.Var]int
		positions  []token.Pos
	}

	extractTypeRequest struct {
		pkg   *packages.Package
		tn    *types.TypeName
		named *types.Named
		opts  *extractorOptions
		idx   *syntaxIndex
	}

	methodBuildRequest struct {
		req        *extractTypeRequest
		refs       *refCollector
		fieldIndex map[*types.Var]int
	}

	typeRefsDocsRequest struct {
		out            *domain.TypeExtract
		req            *extractTypeRequest
		refs           *refCollector
		fieldPositions []token.Pos
		docMethods     []methodDocInput
	}

	memberDocsRequest struct {
		typeDocs       map[types.Object]bool
		fieldDocs      []docRange
		tn             *types.TypeName
		fields         []domain.FieldFacts
		fieldPositions []token.Pos
		methods        []methodDocInput
	}

	methodFactsRequest struct {
		m           methodDecl
		pkg         *packages.Package
		refs        *refCollector
		fieldIndex  map[*types.Var]int
		methodIndex map[*types.Func]int
		opts        *extractorOptions
		fieldCount  int
	}

	sortedMethodsRequest struct {
		fset  *token.FileSet
		named *types.Named
		opts  *extractorOptions
		idx   *syntaxIndex
	}

	skipPosRequest struct {
		fset      *token.FileSet
		generated map[string]bool
		policy    generatedInclusion
		pos       token.Pos
	}

	selectionRequest struct {
		info        *types.Info
		n           *ast.SelectorExpr
		fieldIndex  map[*types.Var]int
		methodIndex map[*types.Func]int
		facts       *domain.MethodFacts
		siblings    map[int]bool
		selfIdx     int
		excludeSelf bool
	}

	walkBodyRequest struct {
		info        *types.Info
		decl        *ast.FuncDecl
		self        *types.Func
		fieldIndex  map[*types.Var]int
		methodIndex map[*types.Func]int
		facts       *domain.MethodFacts
		fieldCount  int
	}

	siblingContext struct {
		siblings    map[int]bool
		selfIdx     int
		excludeSelf bool
	}

	namedRefRequest struct {
		seen     map[string]bool
		self     *types.TypeName
		analyzed map[string]bool
		t        *types.Named
	}

	namedTypeCollect struct {
		out   *domain.PackageExtract
		pkg   *packages.Package
		scope *types.Scope
		opts  *extractorOptions
		idx   *syntaxIndex
	}

	structFieldBuilder struct {
		fields         *[]domain.FieldFacts
		fieldIndex     map[*types.Var]int
		fieldPositions *[]token.Pos
		refs           *refCollector
	}

	typeSpecIndex struct {
		info *types.Info
		idx  *syntaxIndex
		decl *ast.GenDecl
	}

	packageFilterState struct {
		pkgs []*packages.Package
		errs []string
	}

	packageConsider struct {
		path   string
		policy packageErrorPolicy
	}

	extractWorkerState struct {
		job      *extractJob
		extracts []domain.PackageExtract
		opts     extractorOptions
	}

	// Loader is a compiler-backed outbound.FactSource.
	Loader struct{}
)
