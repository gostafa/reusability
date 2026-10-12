// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"testing"

	"github.com/gostafa/reusability"
)

func benchPolicyReport(pkgCount, typesPerPkg int) reusability.Report {
	pkgs := make([]reusability.PackageReport, pkgCount)

	for p := range pkgs {
		types := make([]reusability.TypeReport, typesPerPkg)

		for i := range types {
			metric := typeMetric(float64((p*typesPerPkg+i)%100) / 100)
			if i%7 == 0 {
				metric = naMetric()
			}

			types[i] = reusability.TypeReport{
				Name:        fmt.Sprintf("Type%02d", i),
				Reusability: metric,
			}
		}

		pkgs[p] = reusability.PackageReport{
			Path:  fmt.Sprintf("example.com/m/internal/feature%d/domain", p),
			Types: types,
		}
	}

	return reusability.Report{Packages: pkgs}
}

func BenchmarkEvaluate(b *testing.B) {
	report := benchPolicyReport(100, 25)
	rules := []Rule{
		{Pattern: "**", Min: 0.6},
		{Pattern: "example.com/m/internal/*/domain", Min: 0.7},
		{Pattern: "example.com/m/internal/feature1*/**", Min: 0.8},
	}

	b.ReportAllocs()

	for b.Loop() {
		_ = Evaluate(&report, rules)
	}
}

func BenchmarkMatchPackage(b *testing.B) {
	const path = "example.com/m/internal/features/policy/domain"

	patterns := []string{
		"**",
		"example.com/m/**",
		"example.com/*/internal/**/domain",
		"**/policy/*",
		"example.com/m/internal/features/coupling/domain",
	}

	b.ReportAllocs()

	for b.Loop() {
		for _, pattern := range patterns {
			_ = MatchPackage(pattern, path)
		}
	}
}
