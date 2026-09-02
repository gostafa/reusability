// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
	reusability "github.com/gostafa/reusability/internal/features/reusability/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/workerpool"
)

type (
	// Pipeline orchestrates fact collection and metric computation.
	Pipeline struct {
		analyze func(context.Context, *inbound.Options) (inbound.Result, error)
	}

	pipelineInput = struct {
		facts      typefacts.Collector
		opts       *inbound.Options
		runWorkers func(context.Context, workerpool.RunConfig) error
	}

	packageAnalysisInput = struct {
		pipeline   *pipelineInput
		facts      *tfdomain.ProjectFacts
		calculator *reusability.Service
		pkgID      int
	}

	workerConfigInput = struct {
		pipeline       *pipelineInput
		facts          *tfdomain.ProjectFacts
		calculator     *reusability.Service
		packageResults []inbound.PackageResult
	}
)
