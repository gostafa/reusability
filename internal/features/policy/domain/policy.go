package domain

import (
	"fmt"
	"math"
)

// Rule maps a package import-path glob pattern to a minimum type-level
// reusability threshold. When multiple rules match a package, the strictest
// (highest Min) wins.
type Rule struct {
	// Pattern is a glob against the full import path; * matches one segment,
	// ** matches zero or more segments, using / as the separator.
	Pattern string
	// Min is the minimum acceptable type-level reusability in [0, 1].
	Min float64
}

// DefaultRules returns the recommended baseline when no rules are configured
// (plugin users). A single catch-all rule gates reusability at 0.7.
func DefaultRules() []Rule {
	return []Rule{{Pattern: "**", Min: 0.7}}
}

// Validate checks that every rule has a non-empty pattern and a finite Min in
// [0, 1].
func Validate(rules []Rule) error {
	for i, r := range rules {
		key := fmt.Sprintf("rules[%d]", i)
		if r.Pattern == "" {
			return fmt.Errorf("%s: pattern must be non-empty", key)
		}

		if err := checkFinite(key+".min", r.Min); err != nil {
			return err
		}

		if r.Min < 0 || r.Min > 1 {
			return fmt.Errorf("%s.min: must be in [0, 1], got %g", key, r.Min)
		}
	}

	return nil
}

func checkFinite(key string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s: must be a finite number", key)
	}

	return nil
}
