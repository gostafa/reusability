// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"strconv"
	"strings"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
	"golang.org/x/tools/go/analysis"
)

// New returns a go/analysis Analyzer that loads the module once, evaluates the
// reusability policy, and emits diagnostics for the package under analysis.
func New(settings *Settings) (*analysis.Analyzer, error) {
	resolved := settings.withDefaults()

	err := resolved.validate()
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	active := newRunner(&resolved)

	return &analysis.Analyzer{
		Name: Name,
		Doc:  Doc,
		Run: func(pass *analysis.Pass) (any, error) {
			return active.run(pass)
		},
	}, nil
}

// UnmarshalJSON accepts snake_case tags and remaps kebab-case keys from
// golangci-lint settings so DisallowUnknownFields still applies.
func (settings *Settings) UnmarshalJSON(data []byte) error {
	remapped, err := remapKebabKeys(data)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	err = settings.decodeSettings(remapped)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	return nil
}

func (settings *Settings) decodeSettings(data []byte) error {
	type settingsAlias Settings

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var alias settingsAlias

	err := decoder.Decode(&alias)
	if err != nil {
		return fmt.Errorf(errFmtUnmarshal, err)
	}

	*settings = Settings(alias)

	return nil
}

func (settings *Settings) reusabilityWeights() metrics.ReusabilityWeights {
	weights := metrics.DefaultReusabilityWeights()

	if settings.ReusabilityWeights == nil {
		return weights
	}

	return applyWeightOverrides(&weights, settings.ReusabilityWeights)
}

func applyWeightOverrides(
	weights *metrics.ReusabilityWeights,
	overrides *ReusabilityWeightSettings,
) metrics.ReusabilityWeights {
	// Copy base weights then overlay optional JSON overrides.
	out := *weights

	out.Cohesion = pickFloat(overrides.Cohesion, out.Cohesion)
	out.Coupling = pickFloat(overrides.Coupling, out.Coupling)
	out.Testability = pickFloat(overrides.Testability, out.Testability)
	out.Documentation = pickFloat(overrides.Documentation, out.Documentation)

	return out
}

func pickFloat(override *float64, fallback float64) float64 {
	if override == nil {
		return fallback
	}

	return *override
}

// rules returns the inline policy rules. With no rules configured, the
// recommended defaults apply.
func (settings *Settings) rules() ([]policydomain.Rule, error) {
	if len(settings.Rules) == zero {
		return policydomain.DefaultRules(), nil
	}

	parsed, err := settings.parseRules()
	if err != nil {
		return nil, fmt.Errorf("rules: %w", err)
	}

	return parsed, nil
}

func (settings *Settings) parseRules() ([]policydomain.Rule, error) {
	rules := make([]policydomain.Rule, zero, len(settings.Rules))

	for index := range settings.Rules {
		if settings.Rules[index].Min == nil {
			return nil, fmt.Errorf("parseRules: rules[%d]: %w", index, errRuleMinRequired)
		}

		rules = append(rules, policydomain.Rule{
			Pattern: settings.Rules[index].Pattern,
			Min:     *settings.Rules[index].Min,
		})
	}

	err := policydomain.Validate(rules)
	if err != nil {
		return nil, fmt.Errorf("parseRules: %w", err)
	}

	return rules, nil
}

func (settings *Settings) toConfig() reusability.Config {
	return reusability.Config{
		Directory:          settings.Directory,
		Patterns:           append([]string(nil), settings.Patterns...),
		IncludeTests:       settings.Tests,
		IncludeGenerated:   settings.Generated,
		BuildTags:          append([]string(nil), settings.BuildTags...),
		Workers:            settings.Workers,
		DependencyScope:    reusability.DependencyScope(settings.DependencyScope),
		FieldUsageMode:     reusability.FieldUsageMode(settings.FieldUsage),
		ContinueOnError:    settings.ContinueOnError,
		ReusabilityWeights: settings.reusabilityWeights(),
	}
}

func (settings *Settings) validate() error {
	err := settings.validateModes()
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	err = settings.validateWeightsAndRules()
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	return nil
}

func (settings *Settings) validateModes() error {
	err := validateDependencyScope(settings.DependencyScope)
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	err = validateFieldUsage(settings.FieldUsage)
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	return nil
}

func (settings *Settings) validateWeightsAndRules() error {
	weights := settings.reusabilityWeights()

	err := weights.Validate()
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	rules, err := settings.rules()
	if err != nil {
		return fmt.Errorf(errFmtValidate, err)
	}

	if len(rules) == zero {
		return fmt.Errorf(errFmtValidate, errRuleMinRequired)
	}

	return nil
}

func (settings *Settings) withDefaults() Settings {
	out := *settings

	if len(out.Patterns) == zero {
		out.Patterns = []string{defaultPackagePattern}
	}

	out.DependencyScope = cmp.Or(out.DependencyScope, string(reusability.DependencyScopeModule))
	out.FieldUsage = cmp.Or(out.FieldUsage, string(reusability.FieldUsageDirect))

	return out
}

// Analyze delegates to the adapted function.
func (fn analyzeFunc) Analyze(
	ctx context.Context,
	cfg *reusability.Config,
) (reusability.Report, error) {
	// Forward to the wrapped analyze function and wrap any failure.
	report, err := fn(ctx, cfg)
	if err != nil {
		return reusability.Report{}, fmt.Errorf("Analyze: %w", err)
	}

	return report, nil
}

// computeViolations performs the fallible analysis work before the runner
// caches its result for subsequent package passes.
func computeViolations(
	settings *Settings,
	analyzer reportAnalyzer,
) (map[string][]policydomain.Violation, error) {
	// Load rules, run analysis once, then group violations by package path.
	rules, err := settings.rules()
	if err != nil {
		return nil, fmt.Errorf("reusability policy: %w", err)
	}

	cfg := settings.toConfig()

	report, err := analyzer.Analyze(context.Background(), &cfg)
	if err != nil {
		return nil, fmt.Errorf("reusability analyze: %w", err)
	}

	return groupByPackage(policydomain.Evaluate(&report, rules)), nil
}

func findTypeInFile(file *ast.File, name string) token.Pos {
	var found token.Pos

	ast.Inspect(file, func(node ast.Node) bool {
		if found != token.NoPos {
			return false
		}

		spec, ok := node.(*ast.TypeSpec)

		if !ok || spec.Name == nil || spec.Name.Name != name {
			return true
		}

		found = spec.Name.Pos()

		return false
	})

	return found
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', -one, floatBits)
	}

	return strconv.FormatFloat(value, 'f', two, floatBits)
}

// formatViolation renders one policy violation as a diagnostic message.
func formatViolation(violation *policydomain.Violation) string {
	where := violation.Package + "." + violation.Type + " (type)"

	return fmt.Sprintf("%s: reusability %s is below min %s (rule %s)",
		where,
		formatNumber(violation.Value),
		formatNumber(violation.Threshold),
		violation.Rule,
	)
}

func groupByPackage(violations []policydomain.Violation) map[string][]policydomain.Violation {
	byPkg := make(map[string][]policydomain.Violation, len(violations))

	for index := range violations {
		violation := &violations[index]

		byPkg[violation.Package] = append(byPkg[violation.Package], *violation)
	}

	return byPkg
}

func newRunner(settings *Settings) *runner {
	return &runner{settings: *settings, analyzer: analyzeFunc(reusability.Analyze)}
}

// packagePos returns a position for package-scoped diagnostics: the package
// clause of the first file, or [token.NoPos] when the pass has no files.
func packagePos(pass *analysis.Pass) token.Pos {
	for index := range pass.Files {
		if pass.Files[index] != nil {
			return pass.Files[index].Package
		}
	}

	return token.NoPos
}

func remapKebabKeys(data []byte) ([]byte, error) {
	raw, err := unmarshalSettingsMap(data)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	encoded, err := marshalRemappedKeys(raw)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return encoded, nil
}

func unmarshalSettingsMap(data []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return raw, nil
}

func marshalRemappedKeys(raw map[string]json.RawMessage) ([]byte, error) {
	out := make(map[string]json.RawMessage, len(raw))

	for key := range raw {
		out[strings.ReplaceAll(key, "-", "_")] = raw[key]
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf(errFmtRemap, err)
	}

	return encoded, nil
}

// reportViolations emits one diagnostic per violation for this package pass.
func reportViolations(pass *analysis.Pass, violations []policydomain.Violation) {
	for index := range violations {
		violation := &violations[index]
		pass.Report(analysis.Diagnostic{
			Pos:      violationPos(pass, violation),
			Category: Name,
			Message:  formatViolation(violation),
		})
	}
}

func (runner *runner) load() {
	runner.byPkg, runner.err = computeViolations(&runner.settings, runner.analyzer)
}

func (runner *runner) run(pass *analysis.Pass) (*runResult, error) {
	runner.once.Do(runner.load)

	if runner.err != nil {
		return nil, fmt.Errorf("run: %w", runner.err)
	}

	reportViolations(pass, runner.byPkg[pass.Pkg.Path()])

	return &runResult{}, nil
}

// typePos returns the position of the named type's TypeSpec identifier, or
// the package position when the type is not found in the pass files.
func typePos(pass *analysis.Pass, name string) token.Pos {
	for index := range pass.Files {
		file := pass.Files[index]

		if file == nil {
			continue
		}

		if pos := findTypeInFile(file, name); pos != token.NoPos {
			return pos
		}
	}

	return packagePos(pass)
}

func validateDependencyScope(value string) error {
	switch reusability.DependencyScope(value) {
	case reusability.DependencyScopeProject,
		reusability.DependencyScopeModule,
		reusability.DependencyScopeAll:
		return nil
	default:
		return fmt.Errorf(errFmtQuoted, errInvalidDependencyScope, value)
	}
}

func validateFieldUsage(value string) error {
	switch reusability.FieldUsageMode(value) {
	case reusability.FieldUsageDirect, reusability.FieldUsageTransitive:
		return nil
	default:
		return fmt.Errorf(errFmtQuoted, errInvalidFieldUsage, value)
	}
}

func violationPos(pass *analysis.Pass, violation *policydomain.Violation) token.Pos {
	return typePos(pass, violation.Type)
}
