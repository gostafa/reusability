// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
)

var _ inbound.Analyzer = (*Pipeline)(nil)
