// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

var _ outbound.FactSource = (*Loader)(nil)
