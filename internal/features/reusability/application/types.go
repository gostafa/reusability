// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/reusability/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

type (
	// Result is the reusability domain result for one type.
	Result = domain.Result

	// Calculator evaluates reusability for one type.
	Calculator interface {
		ComputeForType(typeFacts *typefacts.TypeFacts, fieldUsage string) Result
	}

	// Service is the reusability application service.
	Service struct {
		weights metrics.ReusabilityWeights
	}
)
