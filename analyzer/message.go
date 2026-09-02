package analyzer

import (
	"fmt"
	"math"
	"strconv"

	policydomain "github.com/gostafa/reusability/internal/features/policy/domain"
)

// formatViolation renders one policy violation as a diagnostic message.
func formatViolation(v policydomain.Violation) string {
	where := v.Package + " (package)"
	if v.Type != "" {
		where = v.Package + "." + v.Type + " (type)"
	}

	relation := "exceeds max"
	if v.Comparator == policydomain.ComparatorMin {
		relation = "is below min"
	}

	return fmt.Sprintf("%s: %s %s %s %s",
		where, v.Key, formatNumber(v.Value), relation, formatNumber(v.Threshold))
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
