// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	// embed registers the embedded report templates in vars.go.
	_ "embed"
)

const (
	docsDataPlaceholder = "__DOCS_DATA__"
	webDataPlaceholder  = "__REPORT_DATA__"

	replaceOnce = 1
	emptyString = ""
	jsonIndent  = "  "
)
