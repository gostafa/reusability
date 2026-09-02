package domain

import "testing"

// White-box: the guide's direction, boundedness, and column labels mirror
// the renderers' quality map and abbreviations — the guide may never
// contradict how the report actually colors and titles a column.
func TestMetricDocsDirectionMatchesQuality(t *testing.T) {
	t.Parallel()

	for _, d := range MetricDocs() {
		if d.Scope == DocScopeStructural {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: structural direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		if d.Label != abbrev(d.Name) {
			t.Errorf("%s: label = %q, want column heading %q", d.Name, d.Label, abbrev(d.Name))
		}

		q, colored := qualityByMetric[d.Name]
		if !colored {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: uncolored metric direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		want := DirectionHigher
		if q.lowerBetter {
			want = DirectionLower
		}

		if d.Direction != want {
			t.Errorf("%s: direction = %q, want %q", d.Name, d.Direction, want)
		}

		if d.Bounded != q.bounded {
			t.Errorf("%s: bounded = %v, want %v", d.Name, d.Bounded, q.bounded)
		}
	}
}
