// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"github.com/gostafa/reusability/analyzer"
)

type (
	// Plugin is the golangci-lint module plugin for reusability.
	Plugin struct {
		loadMode

		settings analyzer.Settings
	}

	loadMode struct{}
)
