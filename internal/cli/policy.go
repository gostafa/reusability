package cli

import (
	"fmt"
	"strconv"
	"strings"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// override is one CLI policy bound: a condition key and its numeric value.
// Metric keys may be the reported metric ("reusability") or scoped
// ("type.reusability").
type override struct {
	key   string
	value float64
}

// overrideList collects repeated -max / -min flags. It implements flag.Value.
type overrideList struct {
	items []override
}

func (o *overrideList) String() string {
	parts := make([]string, len(o.items))
	for i, ov := range o.items {
		parts[i] = ov.key + "=" + strconv.FormatFloat(ov.value, 'g', -1, 64)
	}

	return strings.Join(parts, ",")
}

func (o *overrideList) Set(value string) error {
	key, number, ok := strings.Cut(value, "=")
	if !ok {
		return fmt.Errorf("expected key=value, got %q", value)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty key in %q", value)
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(number), 64)
	if err != nil {
		return fmt.Errorf("invalid number in %q: %w", value, err)
	}

	o.items = append(o.items, override{key: key, value: parsed})

	return nil
}

// resolvePolicy builds a policy from explicit CLI thresholds only. It returns
// a human-readable source label so the CLI can say which policy ran, and the
// validated policy.
func resolvePolicy(
	maxima, minima overrideList,
) (policydomain.Policy, string, error) {
	if len(maxima.items) == 0 && len(minima.items) == 0 {
		return policydomain.Policy{}, "", fmt.Errorf(
			"no policy thresholds configured; pass -max or -min",
		)
	}

	var policy policydomain.Policy
	for _, ov := range maxima.items {
		if err := policydomain.ApplyOverride(
			&policy,
			ov.key,
			policydomain.ComparatorMax,
			ov.value,
		); err != nil {
			return policydomain.Policy{}, "", err
		}
	}

	for _, ov := range minima.items {
		if err := policydomain.ApplyOverride(
			&policy,
			ov.key,
			policydomain.ComparatorMin,
			ov.value,
		); err != nil {
			return policydomain.Policy{}, "", err
		}
	}

	if err := policydomain.Validate(policy); err != nil {
		return policydomain.Policy{}, "", err
	}

	return policy, "flag thresholds", nil
}
