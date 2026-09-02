package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gostafa/reusability/reusability"
)

// Violation is one type whose reusability fell below the matched rule's
// minimum threshold.
type Violation struct {
	Package   string  // package import path
	Type      string  // type name within the package
	Value     float64 // actual reusability
	Threshold float64 // minimum from the winning rule
	Rule      string  // pattern of the strictest matching rule
}

// Evaluate checks a report against rules and returns violations. Types whose
// reusability metric is not applicable are skipped. When no rules are given,
// DefaultRules applies. Packages and types are already sorted in the report,
// so the result order is deterministic.
func Evaluate(report reusability.Report, rules []Rule) []Violation {
	if len(rules) == 0 {
		rules = DefaultRules()
	}

	var violations []Violation

	for i := range report.Packages {
		pkg := &report.Packages[i]
		min, pattern := strictestRule(pkg.Path, rules)
		if pattern == "" {
			continue
		}

		for j := range pkg.Types {
			typ := &pkg.Types[j]

			if !typ.Reusability.Applicable {
				continue
			}

			value := typ.Reusability.Value

			if value+comparisonTolerance(value, min) < min {
				violations = append(violations, Violation{
					Package:   pkg.Path,
					Type:      typ.Name,
					Value:     value,
					Threshold: min,
					Rule:      pattern,
				})
			}
		}
	}

	return violations
}

func strictestRule(importPath string, rules []Rule) (min float64, pattern string) {
	var matched bool

	for _, rule := range rules {
		if !MatchPackage(rule.Pattern, importPath) {
			continue
		}

		if !matched || rule.Min > min {
			min = rule.Min
			pattern = rule.Pattern
			matched = true
		}
	}

	return min, pattern
}

// comparisonTolerance absorbs floating-point representation noise at a policy
// boundary without hiding a meaningful threshold crossing.
func comparisonTolerance(value, threshold float64) float64 {
	return 1e-12 * max(1, math.Abs(value), math.Abs(threshold))
}

// FormatViolations renders violations as a human-readable summary. The empty
// slice yields the empty string, so callers can print unconditionally.
func FormatViolations(violations []Violation) string {
	if len(violations) == 0 {
		return ""
	}

	var b strings.Builder

	noun := "violations"
	if len(violations) == 1 {
		noun = "violation"
	}

	fmt.Fprintf(&b, "policy: %d %s\n", len(violations), noun)

	for _, v := range violations {
		where := v.Package + "." + v.Type + " (type)"
		fmt.Fprintf(&b, "  %s: reusability %s is below min %s (rule %s)\n",
			where,
			formatNumber(v.Value),
			formatNumber(v.Threshold),
			v.Rule,
		)
	}

	return b.String()
}

// formatNumber prints integers without a fraction and other values with two
// decimals, matching the report's cell formatting.
func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
