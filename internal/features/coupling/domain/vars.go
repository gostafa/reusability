// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/reusability/internal/shared/zerovalue"
)

var _ CouplingGraph = zerovalue.Of[DependencyGraph]()
