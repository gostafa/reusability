// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package bitset

import (
	"github.com/gostafa/reusability/internal/shared/zerovalue"
)

var (
	_ Setter      = zerovalue.Of[FieldSet]()
	_ Intersecter = zerovalue.Of[SmallFieldSet]()
	_ cloner      = zerovalue.Of[FieldSet]()
	_ container   = zerovalue.Of[FieldSet]()
	_ counter     = zerovalue.Of[FieldSet]()
)
