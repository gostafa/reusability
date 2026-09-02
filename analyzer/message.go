package analyzer

import (
	"fmt"
	"math"
	"strconv"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// formatViolation renders one policy violation as a diagnostic message.
func formatViolation(v policydomain.Violation) string {
	where := v.Package + "." + v.Type + " (type)"

	return fmt.Sprintf("%s: reusability %s is below min %s (rule %s)",
		where,
		formatNumber(v.Value),
		formatNumber(v.Threshold),
		v.Rule,
	)
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
