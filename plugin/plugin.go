// Package plugin registers reusability as a golangci-lint Module Plugin.
//
// Blank-import this package from a custom golangci-lint binary (via
// .custom-gcl.yml) to enable the reusability linter. Configuration lives under
// linters.settings.custom.reusability in .golangci.yml.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/reusability/analyzer"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin(analyzer.Name, New)
}

// Plugin adapts the reusability analyzer to golangci-lint's LinterPlugin
// contract.
type Plugin struct {
	loadMode
	settings analyzer.Settings
}

// loadMode supplies the stateless half of register.LinterPlugin separately
// from Plugin's stateful analyzer construction.
type loadMode struct{}

// GetLoadMode requests type information so diagnostics can locate type
// declarations accurately.
func (loadMode) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

var _ register.LinterPlugin = (*Plugin)(nil)

// New constructs the Module Plugin from golangci-lint custom settings.
func New(raw any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[analyzer.Settings](raw)
	if err != nil {
		return nil, err
	}

	return &Plugin{settings: settings}, nil
}

// BuildAnalyzers returns the reusability go/analysis Analyzer.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	a, err := analyzer.New(p.settings)
	if err != nil {
		return nil, err
	}

	return []*analysis.Analyzer{a}, nil
}
