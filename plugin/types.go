// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"golang.org/x/tools/go/analysis"
)

type (
	// Plugin is the golangci-lint module plugin for reusability.
	Plugin struct {
		loadMode

		build func() ([]*analysis.Analyzer, error)
	}

	loadMode struct{}
)
