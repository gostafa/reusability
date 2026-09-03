// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/reusability/domain"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

type (
	// Result is the reusability domain result for one type.
	Result = domain.Result

	// Service is the reusability application service.
	Service = func(*typefacts.TypeFacts, string) Result
)
