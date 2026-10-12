// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"github.com/gostafa/reusability/internal/features/projectanalysis/application"
	"github.com/gostafa/reusability/internal/features/typefacts/adapters/goloader"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
)

// NewAnalyzer returns the production analyzer: go/packages fact extraction
// feeding the metric pipeline.
func NewAnalyzer() application.Pipeline {
	return application.NewPipeline(typefacts.NewServiceFunc(goloader.New()))
}
