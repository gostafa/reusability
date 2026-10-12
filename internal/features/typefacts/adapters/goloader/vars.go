// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import "errors"

var (
	errNoPackagesMatched   = errors.New("no packages matched patterns")
	errNoLoadablePackages  = errors.New("no loadable packages matched patterns")
	errPackageLoadFailures = errors.New("package load errors (use ContinueOnError to skip)")
)
