// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"encoding/json"
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
func New(raw any) (Plugin, error) {
	settings, err := decodePluginSettings(raw)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	return Plugin(analyzerBuilder(&settings)), nil
}

func analyzerBuilder(settings *analyzer.Settings) func() ([]*analysis.Analyzer, error) {
	return func() ([]*analysis.Analyzer, error) {
		analyzerInstance, err := analyzer.New(settings)
		if err != nil {
			return nil, fmt.Errorf("BuildAnalyzers: %w", err)
		}

		return []*analysis.Analyzer{analyzerInstance}, nil
	}
}

func decodePluginSettings(raw any) (analyzer.Settings, error) {
	if raw == nil {
		return analyzer.Settings{}, nil
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return analyzer.Settings{}, fmt.Errorf("marshal settings: %w", err)
	}

	var settings analyzer.Settings

	err = analyzer.UnmarshalSettings(data, &settings)
	if err != nil {
		return analyzer.Settings{}, fmt.Errorf("decode settings: %w", err)
	}

	return settings, nil
}

// BuildAnalyzers returns the reusability go/analysis Analyzer.
func (build Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzers, err := build()
	if err != nil {
		return nil, fmt.Errorf("build analyzers: %w", err)
	}

	return analyzers, nil
}

func (loadMode) value() string {
	return register.LoadModeTypesInfo
}

// GetLoadMode requests type information so diagnostics can locate type
// declarations accurately.
func (Plugin) GetLoadMode() string {
	return loadMode{}.value()
}
