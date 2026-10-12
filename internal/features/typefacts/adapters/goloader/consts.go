// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"time"

	"golang.org/x/tools/go/packages"
)

const (
	zero = 0
	one  = 1

	emptyString = ""

	defaultPackagePattern = "./..."
	testPackageSuffix     = ".test"
	buildTagsPrefix       = "-tags="

	maxShownErrors      = 10
	loadMaxAttempts     = 3
	loadRetryInitial    = 100 * time.Millisecond
	loadRetryMaximum    = 500 * time.Millisecond
	loadRetryBudget     = 2 * time.Second
	loadRetryMultiplier = 2
	loadRetryJitter     = 0.2

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
	errFmtPackageCanceled    = "package load canceled: %w"

	excludeGeneratedFiles generatedInclusion = false
	includeGeneratedFiles generatedInclusion = true

	failOnBrokenPackages packageErrorPolicy = false
	skipBrokenPackages   packageErrorPolicy = true
)
