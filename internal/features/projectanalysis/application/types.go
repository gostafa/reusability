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
		facts      typefacts.Collector
		runWorkers func(context.Context, workerpool.RunConfig) error
	}

	metricNeeds uint8

	packageJob struct {
		facts                 *tfdomain.ProjectFacts
		reusabilityCalculator *reusability.Service
		compute               map[string]bool
		opts                  *inbound.Options
		pkgID                 int
	}

	typeJob struct {
		typeFacts             *tfdomain.TypeFacts
		reusabilityCalculator *reusability.Service
		opts                  *inbound.Options
		needs                 metricNeeds
	}

	assembleJob struct {
		facts                 *tfdomain.ProjectFacts
		reusabilityCalculator *reusability.Service
		compute               map[string]bool
		opts                  *inbound.Options
		runWorkers            func(context.Context, workerpool.RunConfig) error
	}

	analysisRun struct {
		opts       *inbound.Options
		compute    map[string]bool
		calculator *reusability.Service
	}
)
