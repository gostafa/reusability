package analyzer

import (
	"fmt"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// RuleSettings configures one pattern-based reusability rule.
type RuleSettings struct {
	Pattern string   `json:"pattern"`
	Min     *float64 `json:"min"`
}

// rules returns the inline policy rules. With no rules configured, the
// recommended defaults apply.
func (s Settings) rules() ([]policydomain.Rule, error) {
	if len(s.Rules) == 0 {
		return policydomain.DefaultRules(), nil
	}

	rules := make([]policydomain.Rule, len(s.Rules))
	for i, r := range s.Rules {
		if r.Min == nil {
			return nil, errRuleMinRequired(i)
		}

		rules[i] = policydomain.Rule{Pattern: r.Pattern, Min: *r.Min}
	}

	if err := policydomain.Validate(rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func errRuleMinRequired(index int) error {
	return fmt.Errorf("rules[%d]: min is required", index)
}
