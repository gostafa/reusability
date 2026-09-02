package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

// The canonical keys for the structural conditions. Metric conditions are keyed
// by their metric name (amc, cbo, distance, ...), optionally prefixed with
// "package." or "type." in CLI overrides for scope-specific bounds.
const (
	KeyTypes           = "types"            // named types per package
	KeyFuncs           = "funcs"            // declared functions and methods per package
	KeyExportedFuncs   = "exported_funcs"   // exported functions and methods per package
	KeyUnexportedFuncs = "unexported_funcs" // unexported functions and methods per package
	KeyVars            = "vars"             // top-level variable names per package
	KeyConsts          = "consts"           // top-level constant names per package
	KeyAfferent        = "afferent"         // incoming coupling (Ca) per package
	KeyEfferent        = "efferent"         // outgoing coupling (Ce) per package
	KeyFields          = "fields"           // struct fields per type
	KeyMethods         = "methods"          // declared methods per type
	KeyFuncLines       = "funcs.lines"      // source lines per function or method
	KeyFuncCyclomatic  = "funcs.cyclomatic" // cyclomatic complexity per function or method
)

// Limit is a single condition on one field: an optional upper bound, an
// optional lower bound, or both. The zero value imposes no constraint.
type Limit struct {
	// Max is the upper bound when HasMax is true.
	Max float64
	// HasMax reports whether Max is set.
	HasMax bool
	// Min is the lower bound when HasMin is true.
	Min float64
	// HasMin reports whether Min is set.
	HasMin bool
}

// validate reports whether the bounds are finite and mutually consistent.
func (l Limit) validate(key string) error {
	if err := checkFinite(key, "max", l.HasMax, l.Max); err != nil {
		return err
	}

	if err := checkFinite(key, "min", l.HasMin, l.Min); err != nil {
		return err
	}

	return checkOrder(key, l)
}

// checkFinite reports an error when a set bound is not a finite number. side
// names the bound ("max" or "min") for the message.
func checkFinite(key, side string, has bool, value float64) error {
	if has && !finite(value) {
		return fmt.Errorf("%s: %s must be a finite number", key, side)
	}

	return nil
}

// checkOrder reports an error when both bounds are set and min exceeds max.
func checkOrder(key string, l Limit) error {
	if l.HasMax && l.HasMin && l.Min > l.Max {
		return fmt.Errorf("%s: min %g exceeds max %g", key, l.Min, l.Max)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// PackageLimits holds the conditions evaluated once per package.
type PackageLimits struct {
	Types           Limit // budget on named types per package
	Funcs           FuncLimits
	ExportedFuncs   Limit // budget on exported functions and methods
	UnexportedFuncs Limit // budget on unexported functions and methods
	Vars            Limit // budget on top-level variable names
	Consts          Limit // budget on top-level constant names
	Afferent        Limit // budget on incoming coupling (Ca)
	Efferent        Limit // budget on outgoing coupling (Ce)
}

// TypeLimits holds the conditions evaluated once per named type.
type TypeLimits struct {
	Fields  Limit // budget on struct fields per type
	Methods Limit // budget on declared methods per type
}

// FuncLimits holds conditions evaluated for a function set. Count applies to
// the set size; Lines and Cyclomatic apply to each function or method.
type FuncLimits struct {
	Count      Limit // budget on functions or methods in the parent entity
	Lines      Limit // budget on source lines per function or method
	Cyclomatic Limit // budget on cyclomatic complexity per function or method
}

// Policy is a complete set of conditions to enforce against a report.
// Metrics is the legacy/global metric map: a condition there applies to any
// scope where that metric is present. TypeMetrics is the type-scoped map; a
// scoped condition replaces a same-named global condition. Only the reported
// metric is gateable, and only where it is present and applicable.
type Policy struct {
	Package     PackageLimits    // conditions evaluated once per package
	Type        TypeLimits       // conditions evaluated once per named type
	Funcs       FuncLimits       // conditions evaluated for functions and methods
	Metrics     map[string]Limit // legacy/global metric conditions, keyed by metric name
	TypeMetrics map[string]Limit // type metric conditions, keyed by metric name
}

// DefaultPolicy returns the recommended strict baseline. It gates the robust,
// low-false-positive fields; afferent coupling is left unconstrained, so opt
// into it explicitly.
func DefaultPolicy() Policy {
	return Policy{
		Package: PackageLimits{
			Types: Limit{Max: 12, HasMax: true},
			Funcs: FuncLimits{
				Count: Limit{Max: 30, HasMax: true},
			},
			Efferent: Limit{Max: 10, HasMax: true},
		},
		Type: TypeLimits{
			Fields:  Limit{Max: 8, HasMax: true},
			Methods: Limit{Max: 10, HasMax: true},
		},
		Funcs: FuncLimits{
			Lines:      Limit{Max: 80, HasMax: true},
			Cyclomatic: Limit{Max: 10, HasMax: true},
		},
		TypeMetrics: map[string]Limit{
			metrics.MetricReusability: {Min: 0.7, HasMin: true},
		},
	}
}

// Validate checks that every structural condition is finite and consistent and
// that every metric key names a real metric.
func Validate(p Policy) error {
	if err := validateStructural(p); err != nil {
		return err
	}

	return validateMetrics(p)
}

// validateStructural checks every structural condition for finiteness and
// consistency.
func validateStructural(p Policy) error {
	structural := []struct {
		key   string
		limit Limit
	}{
		{KeyTypes, p.Package.Types},
		{KeyFuncs, p.Package.Funcs.Count},
		{KeyExportedFuncs, p.Package.ExportedFuncs},
		{KeyUnexportedFuncs, p.Package.UnexportedFuncs},
		{KeyVars, p.Package.Vars},
		{KeyConsts, p.Package.Consts},
		{KeyAfferent, p.Package.Afferent},
		{KeyEfferent, p.Package.Efferent},
		{KeyFields, p.Type.Fields},
		{KeyMethods, p.Type.Methods},
		{KeyFuncLines, p.Funcs.Lines},
		{KeyFuncCyclomatic, p.Funcs.Cyclomatic},
		{"package." + KeyFuncLines, p.Package.Funcs.Lines},
		{"package." + KeyFuncCyclomatic, p.Package.Funcs.Cyclomatic},
	}

	for _, s := range structural {
		if err := s.limit.validate(s.key); err != nil {
			return err
		}
	}

	return nil
}

// validateMetrics checks that every metric key names a real metric in its
// configured scope and that its bounds are finite and consistent.
func validateMetrics(p Policy) error {
	if err := validateMetricMap("metric", "", p.Metrics, knownMetrics()); err != nil {
		return err
	}

	return validateMetricMap("type metric", "type.", p.TypeMetrics, knownMetrics())
}

// validateMetricMap visits names in sorted order so the first reported error is
// deterministic.
func validateMetricMap(
	label, keyPrefix string,
	metricLimits map[string]Limit,
	known map[string]bool,
) error {
	names := make([]string, 0, len(metricLimits))
	for name := range metricLimits {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		if !known[name] {
			return fmt.Errorf("unknown policy metric %q", name)
		}

		if err := metricLimits[name].validate(keyPrefix + name); err != nil {
			return err
		}
	}

	return nil
}

// ApplyOverride sets one bound (max or min) on the condition named by key,
// which may be a structural key, a legacy/global metric name, or a scoped key
// in the form "package.<structural>" or "type.<key>". Metric conditions are
// created on demand. An unknown key is an error.
func ApplyOverride(p *Policy, key string, cmp Comparator, value float64) error {
	if ptr, ok := structuralLimit(p, key); ok {
		setBound(ptr, cmp, value)

		return nil
	}

	if scope, name, ok := strings.Cut(key, "."); ok {
		if ptr, ok := scopedStructuralLimit(p, scope, name); ok {
			setBound(ptr, cmp, value)

			return nil
		}

		switch scope {
		case "type":
			if !knownMetrics()[name] {
				return fmt.Errorf("unknown policy metric %q", name)
			}

			setMetricBound(&p.TypeMetrics, name, cmp, value)

			return nil
		}
	}

	if knownMetrics()[key] {
		setMetricBound(&p.Metrics, key, cmp, value)

		return nil
	}

	return fmt.Errorf("unknown policy key %q", key)
}

// MetricNames returns, sorted, every metric name the policy constrains. The
// CLI unions these into the selected metric set so gated metrics are computed
// and rendered — a metric absent from the report is never checked.
func MetricNames(p Policy) []string {
	seen := make(map[string]bool, len(p.Metrics)+len(p.TypeMetrics))
	addMetricNames(seen, p.Metrics)
	addMetricNames(seen, p.TypeMetrics)

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func addMetricNames(names map[string]bool, limits map[string]Limit) {
	for name, limit := range limits {
		if hasBounds(limit) {
			names[name] = true
		}
	}
}

// structuralLimit returns an addressable pointer to the structural condition
// named by key, or false when key is not a structural key.
func structuralLimit(p *Policy, key string) (*Limit, bool) {
	switch key {
	case KeyTypes:
		return &p.Package.Types, true
	case KeyFuncs:
		return &p.Package.Funcs.Count, true
	case KeyExportedFuncs:
		return &p.Package.ExportedFuncs, true
	case KeyUnexportedFuncs:
		return &p.Package.UnexportedFuncs, true
	case KeyVars:
		return &p.Package.Vars, true
	case KeyConsts:
		return &p.Package.Consts, true
	case KeyAfferent:
		return &p.Package.Afferent, true
	case KeyEfferent:
		return &p.Package.Efferent, true
	case KeyFields:
		return &p.Type.Fields, true
	case KeyMethods:
		return &p.Type.Methods, true
	case KeyFuncLines:
		return &p.Funcs.Lines, true
	case KeyFuncCyclomatic:
		return &p.Funcs.Cyclomatic, true
	}

	return nil, false
}

func scopedStructuralLimit(p *Policy, scope, key string) (*Limit, bool) {
	switch scope {
	case "package":
		return packageStructuralLimit(p, key)
	case "type":
		return typeStructuralLimit(p, key)
	default:
		return nil, false
	}
}

func packageStructuralLimit(p *Policy, key string) (*Limit, bool) {
	switch key {
	case KeyTypes:
		return &p.Package.Types, true
	case KeyFuncs:
		return &p.Package.Funcs.Count, true
	case KeyExportedFuncs:
		return &p.Package.ExportedFuncs, true
	case KeyUnexportedFuncs:
		return &p.Package.UnexportedFuncs, true
	case KeyVars:
		return &p.Package.Vars, true
	case KeyConsts:
		return &p.Package.Consts, true
	case KeyAfferent:
		return &p.Package.Afferent, true
	case KeyEfferent:
		return &p.Package.Efferent, true
	case KeyFuncLines:
		return &p.Package.Funcs.Lines, true
	case KeyFuncCyclomatic:
		return &p.Package.Funcs.Cyclomatic, true
	default:
		return nil, false
	}
}

func typeStructuralLimit(p *Policy, key string) (*Limit, bool) {
	switch key {
	case KeyFields:
		return &p.Type.Fields, true
	case KeyMethods:
		return &p.Type.Methods, true
	case KeyFuncLines:
		return &p.Funcs.Lines, true
	case KeyFuncCyclomatic:
		return &p.Funcs.Cyclomatic, true
	default:
		return nil, false
	}
}

func setBound(limit *Limit, cmp Comparator, value float64) {
	switch cmp {
	case ComparatorMax:
		limit.Max, limit.HasMax = value, true
	case ComparatorMin:
		limit.Min, limit.HasMin = value, true
	}
}

func setMetricBound(metricLimits *map[string]Limit, name string, cmp Comparator, value float64) {
	if *metricLimits == nil {
		*metricLimits = make(map[string]Limit)
	}

	limit := (*metricLimits)[name]
	setBound(&limit, cmp, value)
	(*metricLimits)[name] = limit
}

func typeMetricLimit(policy Policy, name string) (Limit, bool) {
	return scopedMetricLimit(policy.Metrics, policy.TypeMetrics, name)
}

func scopedMetricLimit(global, scoped map[string]Limit, name string) (Limit, bool) {
	if limit, ok := scoped[name]; ok {
		return limit, hasBounds(limit)
	}

	limit, ok := global[name]

	return limit, ok && hasBounds(limit)
}

func hasBounds(limit Limit) bool {
	return limit.HasMax || limit.HasMin
}

// knownMetrics is the set of selectable metric names, sourced from the facade
// so the policy stays in step with the analyzer.
func knownMetrics() map[string]bool {
	all := reusability.AllMetrics()

	set := make(map[string]bool, len(all))
	for _, name := range all {
		set[string(name)] = true
	}

	return set
}

func metricSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	return set
}
