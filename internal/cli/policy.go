package cli

import (
	"fmt"
	"strconv"
	"strings"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// ruleSpec is one CLI policy rule: a package-path pattern and its minimum
// reusability threshold.
type ruleSpec struct {
	pattern string
	min     float64
}

// ruleList collects repeated -rule flags. It implements flag.Value.
type ruleList struct {
	items []ruleSpec
}

func (r *ruleList) String() string {
	parts := make([]string, len(r.items))
	for i, spec := range r.items {
		parts[i] = spec.pattern + ":" + strconv.FormatFloat(spec.min, 'g', -1, 64)
	}

	return strings.Join(parts, ",")
}

func (r *ruleList) Set(value string) error {
	pattern, number, ok := strings.Cut(value, ":")
	if !ok {
		return fmt.Errorf("expected pattern:min, got %q", value)
	}

	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return fmt.Errorf("empty pattern in %q", value)
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(number), 64)
	if err != nil {
		return fmt.Errorf("invalid number in %q: %w", value, err)
	}

	r.items = append(r.items, ruleSpec{pattern: pattern, min: parsed})

	return nil
}

// resolvePolicy builds rules from explicit CLI -rule flags. It returns a
// human-readable source label and the validated rules.
func resolvePolicy(rules ruleList) ([]policydomain.Rule, string, error) {
	if len(rules.items) == 0 {
		return nil, "", fmt.Errorf(
			"no policy rules configured; pass at least one -rule=pattern:min with -check",
		)
	}

	out := make([]policydomain.Rule, len(rules.items))
	for i, spec := range rules.items {
		out[i] = policydomain.Rule{Pattern: spec.pattern, Min: spec.min}
	}

	if err := policydomain.Validate(out); err != nil {
		return nil, "", err
	}

	return out, "flag rules", nil
}
