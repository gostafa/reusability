package analyzer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// PackageSettings configures limits evaluated once per package.
type PackageSettings struct {
	Types           *LimitSettings `json:"types"`
	Funcs           *FuncSettings  `json:"funcs"`
	ExportedFuncs   *LimitSettings `json:"exported_funcs"`
	UnexportedFuncs *LimitSettings `json:"unexported_funcs"`
	Vars            *LimitSettings `json:"vars"`
	Consts          *LimitSettings `json:"consts"`
	Afferent        *LimitSettings `json:"afferent"`
	Efferent        *LimitSettings `json:"efferent"`
}

// TypeSettings configures limits evaluated once per named type.
type TypeSettings struct {
	Fields  *LimitSettings           `json:"fields"`
	Methods *LimitSettings           `json:"methods"`
	Metrics map[string]LimitSettings `json:"metrics"`
}

// LimitSettings is an optional lower bound, upper bound, or both. When decoded
// from golangci-lint settings, a bare number is shorthand for a maximum.
type LimitSettings struct {
	Max *float64 `json:"max"`
	Min *float64 `json:"min"`
}

// FuncSettings configures a function set. A bare number is shorthand for a
// count maximum; lines and cyclomatic apply to each function or method.
type FuncSettings struct {
	Max        *float64       `json:"max"`
	Min        *float64       `json:"min"`
	Lines      *LimitSettings `json:"lines"`
	Cyclomatic *LimitSettings `json:"cyclomatic"`
}

// UnmarshalJSON accepts a bare numeric count maximum or a strict object with
// optional max/min count bounds and per-function lines/cyclomatic limits.
func (f *FuncSettings) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return errors.New("funcs must be a number or a {min, max, lines, cyclomatic} mapping")
	}

	var scalar float64
	if err := json.Unmarshal(trimmed, &scalar); err == nil {
		f.Max = &scalar
		f.Min = nil
		f.Lines = nil
		f.Cyclomatic = nil

		return nil
	}

	var fields struct {
		Max        *float64       `json:"max"`
		Min        *float64       `json:"min"`
		Lines      *LimitSettings `json:"lines"`
		Cyclomatic *LimitSettings `json:"cyclomatic"`
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fields); err != nil {
		return fmt.Errorf(
			"funcs must be a number or a {min, max, lines, cyclomatic} mapping: %w",
			err,
		)
	}

	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}

	if fields.Max == nil && fields.Min == nil && fields.Lines == nil && fields.Cyclomatic == nil {
		return errors.New("funcs must set max, min, lines, and/or cyclomatic")
	}

	f.Max = fields.Max
	f.Min = fields.Min
	f.Lines = fields.Lines
	f.Cyclomatic = fields.Cyclomatic

	return nil
}

// UnmarshalJSON accepts either a bare numeric maximum or a strict
// {"min": ..., "max": ...} object.
func (l *LimitSettings) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return errors.New("limit must be a number or a {min, max} mapping")
	}

	var scalar float64
	if err := json.Unmarshal(trimmed, &scalar); err == nil {
		l.Max = &scalar
		l.Min = nil

		return nil
	}

	var bounds struct {
		Max *float64 `json:"max"`
		Min *float64 `json:"min"`
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&bounds); err != nil {
		return fmt.Errorf("limit must be a number or a {min, max} mapping: %w", err)
	}

	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}

	if bounds.Max == nil && bounds.Min == nil {
		return errors.New("limit must set max and/or min")
	}

	l.Max = bounds.Max
	l.Min = bounds.Min

	return nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("limit must contain exactly one JSON value")
		}

		return fmt.Errorf("decoding limit: %w", err)
	}

	return nil
}

// policy returns the inline policy. With no policy keys configured, the
// recommended defaults apply. It never reads or discovers a policy file.
func (s Settings) policy() (policydomain.Policy, error) {
	if s.Package == nil && s.Type == nil && s.Funcs == nil && s.Metrics == nil {
		return policydomain.DefaultPolicy(), nil
	}

	policy := policydomain.Policy{
		Metrics:     make(map[string]policydomain.Limit, len(s.Metrics)),
		TypeMetrics: make(map[string]policydomain.Limit),
	}

	if err := applyPackageSettings(&policy, s.Package); err != nil {
		return policydomain.Policy{}, err
	}

	if err := applyTypeSettings(&policy, s.Type); err != nil {
		return policydomain.Policy{}, err
	}

	if err := applyRootFuncSettings(&policy, s.Funcs); err != nil {
		return policydomain.Policy{}, err
	}

	if err := copyLimitSettings(policy.Metrics, s.Metrics); err != nil {
		return policydomain.Policy{}, err
	}

	if err := policydomain.Validate(policy); err != nil {
		return policydomain.Policy{}, err
	}

	return policy, nil
}

func applyPackageSettings(policy *policydomain.Policy, settings *PackageSettings) error {
	if settings == nil {
		return nil
	}

	applyLimitSettings([]limitSettingBinding{
		{settings: settings.Types, destination: &policy.Package.Types},
		{settings: settings.ExportedFuncs, destination: &policy.Package.ExportedFuncs},
		{settings: settings.UnexportedFuncs, destination: &policy.Package.UnexportedFuncs},
		{settings: settings.Vars, destination: &policy.Package.Vars},
		{settings: settings.Consts, destination: &policy.Package.Consts},
		{settings: settings.Afferent, destination: &policy.Package.Afferent},
		{settings: settings.Efferent, destination: &policy.Package.Efferent},
	})
	if settings.Funcs != nil {
		policy.Package.Funcs = settings.Funcs.toFuncLimits()
	}

	return nil
}

func applyTypeSettings(policy *policydomain.Policy, settings *TypeSettings) error {
	if settings == nil {
		return nil
	}

	applyLimitSettings([]limitSettingBinding{
		{settings: settings.Fields, destination: &policy.Type.Fields},
		{settings: settings.Methods, destination: &policy.Type.Methods},
	})
	policy.TypeMetrics = make(map[string]policydomain.Limit, len(settings.Metrics))

	return copyLimitSettings(policy.TypeMetrics, settings.Metrics)
}

func applyRootFuncSettings(policy *policydomain.Policy, settings *FuncSettings) error {
	if settings == nil {
		return nil
	}

	limits := settings.toFuncLimits()
	if limits.Count.HasMax || limits.Count.HasMin {
		return errors.New("funcs count limit belongs under package.funcs")
	}

	policy.Funcs = limits

	return nil
}

type limitSettingBinding struct {
	settings    *LimitSettings
	destination *policydomain.Limit
}

func applyLimitSettings(bindings []limitSettingBinding) {
	for _, binding := range bindings {
		*binding.destination = binding.settings.toLimit()
	}
}

func (l *LimitSettings) toLimit() policydomain.Limit {
	if l == nil {
		return policydomain.Limit{}
	}

	limit := policydomain.Limit{}
	if l.Max != nil {
		limit.Max = *l.Max
		limit.HasMax = true
	}
	if l.Min != nil {
		limit.Min = *l.Min
		limit.HasMin = true
	}

	return limit
}

func (f *FuncSettings) toFuncLimits() policydomain.FuncLimits {
	if f == nil {
		return policydomain.FuncLimits{}
	}

	return policydomain.FuncLimits{
		Count:      limitFromBounds(f.Max, f.Min),
		Lines:      f.Lines.toLimit(),
		Cyclomatic: f.Cyclomatic.toLimit(),
	}
}

func limitFromBounds(maximum, minimum *float64) policydomain.Limit {
	limit := policydomain.Limit{}
	if maximum != nil {
		limit.Max = *maximum
		limit.HasMax = true
	}
	if minimum != nil {
		limit.Min = *minimum
		limit.HasMin = true
	}

	return limit
}

func copyLimitSettings(
	destination map[string]policydomain.Limit,
	source map[string]LimitSettings,
) error {
	for name, settings := range source {
		limit := settings.toLimit()
		if !limit.HasMax && !limit.HasMin {
			return fmt.Errorf("%s: limit must set max and/or min", name)
		}

		destination[name] = limit
	}

	return nil
}
