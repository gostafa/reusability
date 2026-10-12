// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gostafa/reusability"
	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

func benchReport(pkgCount, typesPerPkg int) reusability.Report {
	pkgs := make([]reusability.PackageReport, pkgCount)

	for p := range pkgs {
		types := make([]reusability.TypeReport, typesPerPkg)

		for i := range types {
			types[i] = reusability.TypeReport{
				Name: fmt.Sprintf("Type%02d", i),
				Reusability: metrics.MetricResult{
					Name:       metrics.MetricReusability,
					Scope:      metrics.ScopeType,
					Value:      float64((p*typesPerPkg+i)%100) / 100,
					Applicable: i%7 != 0,
				},
			}
		}

		pkgs[p] = reusability.PackageReport{
			Path:  fmt.Sprintf("example.com/m/pkg%03d", p),
			Types: types,
		}
	}

	return reusability.Report{
		SchemaVersion: reusability.SchemaVersion,
		Tool:          reusability.ToolInfo{Name: reusability.ToolName(), Version: "bench"},
		Module:        "example.com/m",
		Packages:      pkgs,
	}
}

func BenchmarkWrite(b *testing.B) {
	report := benchReport(100, 25)

	for _, format := range []domain.Format{
		domain.FormatText,
		domain.FormatJSON,
		domain.FormatCSV,
		domain.FormatWeb,
	} {
		b.Run(string(format), func(b *testing.B) {
			var buf bytes.Buffer

			req := WriteRequest{Sink: bufferSink(&buf), Format: format, Report: report}

			b.ReportAllocs()

			for b.Loop() {
				buf.Reset()

				err := Write(&req)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
