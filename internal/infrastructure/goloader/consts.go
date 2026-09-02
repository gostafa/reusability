// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"golang.org/x/tools/go/packages"
)

const (
	zero = 0
	one  = 1

	emptyString = ""

	defaultPackagePattern = "./..."
	testPackageSuffix     = ".test"
	buildTagsPrefix       = "-tags="

	maxShownErrors = 10

	loadMode = packages.NeedName |
		packages.NeedModule |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	parentDir = ".."

	errWrapWithValue = "%w: %v"

	errFmtInvokePackagesLoad = "invoke packages load: %w"

	excludeGeneratedFiles generatedInclusion = false
	includeGeneratedFiles generatedInclusion = true

	failOnBrokenPackages packageErrorPolicy = false
	skipBrokenPackages   packageErrorPolicy = true
)
