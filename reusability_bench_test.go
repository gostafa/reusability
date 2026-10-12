// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability

import (
	"context"
	"testing"
)

// BenchmarkAnalyze measures the full public pipeline: package loading, fact
// extraction, metric computation, and report assembly over the fixture module.
func BenchmarkAnalyze(b *testing.B) {
	ctx := context.Background()
	cfg := Config{Directory: "testdata/fixture"}

	b.ReportAllocs()

	for b.Loop() {
		_, err := Analyze(ctx, &cfg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
