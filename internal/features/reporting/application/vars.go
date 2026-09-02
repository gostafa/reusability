// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	// embed is required for the //go:embed directives in vars.go.
	_ "embed"
	"errors"
)

var (
	// docsTemplate is the self-contained metrics guide page.
	//
	//go:embed web_docs_template.html
	docsTemplate string

	// webTemplate is the self-contained HTML report page.
	//
	//go:embed web_template.html
	webTemplate string

	errUnknownFormat      = errors.New("unknown report format")
	errShortWrite         = errors.New("short write")
	errMissingPlaceholder = errors.New("template is missing placeholder")
)
