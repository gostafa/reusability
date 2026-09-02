// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/reusability/analyzer"
	"golang.org/x/tools/go/analysis"
)

func registerModule() int {
	register.Plugin(analyzer.Name, func(conf any) (register.LinterPlugin, error) {
		pluginInstance, err := New(conf)
		if err != nil {
			return nil, fmt.Errorf("registerModule: %w", err)
		}

		return pluginInstance, nil
	})

	return registerDone
}

// New constructs the Module Plugin from golangci-lint custom settings.
func New(raw any) (*Plugin, error) {
	settings, err := register.DecodeSettings[analyzer.Settings](raw)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	return &Plugin{settings: settings}, nil
}

// BuildAnalyzers returns the reusability go/analysis Analyzer.
func (plugin *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzerInstance, err := analyzer.New(&plugin.settings)
	if err != nil {
		return nil, fmt.Errorf("BuildAnalyzers: %w", err)
	}

	return []*analysis.Analyzer{analyzerInstance}, nil
}

// GetLoadMode requests type information so diagnostics can locate type
// declarations accurately.
func (loadMode) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
