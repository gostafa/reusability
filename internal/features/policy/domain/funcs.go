// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gostafa/reusability/reusability"
)

// DefaultRules returns the recommended baseline when no rules are configured
// (plugin users). A single catch-all rule gates reusability at 0.7.
func DefaultRules() []Rule {
	return []Rule{{Pattern: doubleStar, Min: defaultMinReusability}}
}

// Evaluate checks a report against rules and returns violations. Types whose
// reusability metric is not applicable are skipped. When no rules are given,
// DefaultRules applies. Packages and types are already sorted in the report,
// so the result order is deterministic.
func Evaluate(report *reusability.Report, rules []Rule) []Violation {
	if len(rules) == zero {
		rules = DefaultRules()
	}

	violations := make([]Violation, zero, countReportTypes(report))

	for index := range report.Packages {
		violations = append(violations, evaluatePackage(&report.Packages[index], rules)...)
	}

	return violations
}

// FormatViolations renders violations as a human-readable summary. The empty
// slice yields the empty string, so callers can print unconditionally.
func FormatViolations(violations []Violation) string {
	if len(violations) == zero {
		return emptyString
	}

	var builder strings.Builder

	err := writeViolationHeader(&builder, len(violations))
	if err != nil {
		return emptyString
	}

	err = writeViolationLines(&builder, violations)
	if err != nil {
		return emptyString
	}

	return builder.String()
}

// MatchPackage reports whether pattern matches importPath. * matches exactly
// one path segment; ** matches zero or more segments. Matching is against
// the full import path using / as the separator.
func MatchPackage(pattern, importPath string) bool {
	return matchSegments(splitPattern(pattern), strings.Split(importPath, pathSep))
}

// Validate checks that every rule has a non-empty pattern and a finite Min in
// [0, 1].
func Validate(rules []Rule) error {
	for index := range rules {
		err := validateRule(fmt.Sprintf("rules[%d]", index), rules[index])
		if err != nil {
			return fmt.Errorf("Validate: %w", err)
		}
	}

	return nil
}

func checkFinite(key string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, zero) {
		return fmt.Errorf(errFmtKeyed, key, errNotFinite)
	}

	return nil
}

func comparisonTolerance(value, threshold float64) float64 {
	return comparisonEps * max(one, math.Abs(value), math.Abs(threshold))
}

func countReportTypes(report *reusability.Report) int {
	total := zero

	for index := range report.Packages {
		total += len(report.Packages[index].Types)
	}

	return total
}

func evaluatePackage(pkg *reusability.PackageReport, rules []Rule) []Violation {
	threshold, pattern := strictestRule(pkg.Path, rules)

	if pattern == emptyString {
		return nil
	}

	var violations []Violation

	for index := range pkg.Types {
		if violation, ok := violationForType(&typeGate{
			pkgPath: pkg.Path, typ: &pkg.Types[index], threshold: threshold, pattern: pattern,
		}); ok {
			violations = append(violations, violation)
		}
	}

	return violations
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', -one, floatBits)
	}

	return strconv.FormatFloat(value, 'f', two, floatBits)
}

func matchDoubleStar(pattern, path []string, pos matchPos) bool {
	pos.pi++

	if pos.pi == len(pattern) {
		return true
	}

	for pos.si <= len(path) {
		if matchFrom(pattern, path, pos) {
			return true
		}

		pos.si++
	}

	return false
}

func matchFrom(pattern, path []string, pos matchPos) bool {
	for pos.pi < len(pattern) {
		if pattern[pos.pi] == doubleStar {
			return matchDoubleStar(pattern, path, pos)
		}

		if !matchOne(pattern, path, &pos) {
			return false
		}
	}

	return pos.si == len(path)
}

func matchOne(pattern, path []string, pos *matchPos) bool {
	if pos.si >= len(path) {
		return false
	}

	if pattern[pos.pi] != star && pattern[pos.pi] != path[pos.si] {
		return false
	}

	pos.pi++

	pos.si++

	return true
}

func matchSegments(pattern, path []string) bool {
	return matchFrom(pattern, path, matchPos{})
}

func violationForType(gate *typeGate) (Violation, bool) {
	if !gate.typ.Reusability.Applicable {
		return Violation{}, false
	}

	value := gate.typ.Reusability.Value

	if value+comparisonTolerance(value, gate.threshold) >= gate.threshold {
		return Violation{}, false
	}

	return Violation{
		Package:   gate.pkgPath,
		Type:      gate.typ.Name,
		Value:     value,
		Threshold: gate.threshold,
		Rule:      gate.pattern,
	}, true
}

func splitPattern(pattern string) []string {
	if pattern == emptyString {
		return nil
	}

	return strings.Split(pattern, pathSep)
}

func strictestRule(importPath string, rules []Rule) (threshold float64, pattern string) {
	var matched bool

	for index := range rules {
		candidate := ruleCandidate{
			importPath: importPath,
			rule:       rules[index],
			threshold:  threshold,
			pattern:    pattern,
			matched:    matched,
		}

		considerRule(&candidate)

		threshold, pattern, matched = candidate.threshold, candidate.pattern, candidate.matched
	}

	return threshold, pattern
}

func considerRule(candidate *ruleCandidate) {
	if !MatchPackage(candidate.rule.Pattern, candidate.importPath) {
		return
	}

	if !candidate.matched || candidate.rule.Min > candidate.threshold {
		candidate.threshold = candidate.rule.Min
		candidate.pattern = candidate.rule.Pattern
		candidate.matched = true
	}
}

func validateRule(key string, rule Rule) error {
	if rule.Pattern == emptyString {
		return fmt.Errorf(errFmtKeyed, key, errPatternEmpty)
	}

	err := checkFinite(key+".min", rule.Min)
	if err != nil {
		return fmt.Errorf("validateRule: %w", err)
	}

	if rule.Min < zero || rule.Min > one {
		return fmt.Errorf("%s.min: %w (got %g)", key, errMinOutOfRange, rule.Min)
	}

	return nil
}

func writef(builder *strings.Builder, format string, args ...any) error {
	written, err := fmt.Fprintf(builder, format, args...)
	if err != nil {
		return fmt.Errorf(errFmtWrite, err)
	}

	if written < zero {
		return fmt.Errorf(errFmtWrite, errNegativeWrite)
	}

	return nil
}

func writeViolationHeader(builder *strings.Builder, count int) error {
	noun := "violations"

	if count == one {
		noun = "violation"
	}

	err := writef(builder, "policy: %d %s\n", count, noun)
	if err != nil {
		return fmt.Errorf("writeViolationHeader: %w", err)
	}

	return nil
}

func writeViolationLines(builder *strings.Builder, violations []Violation) error {
	for index := range violations {
		violation := &violations[index]
		where := violation.Package + "." + violation.Type + " (type)"

		err := writef(builder, "  %s: reusability %s is below min %s (rule %s)\n",
			where,
			formatNumber(violation.Value),
			formatNumber(violation.Threshold),
			violation.Rule,
		)
		if err != nil {
			return fmt.Errorf("writeViolationLines: %w", err)
		}
	}

	return nil
}
