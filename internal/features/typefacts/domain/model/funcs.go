// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"fmt"

	"github.com/gostafa/reusability/internal/shared/bitset"
)

// String summarizes the method facts for debugging.
func (method *MethodFacts) String() string {
	return fmt.Sprintf(
		"method %q (exported %v) at %v: uses %d fields, branches %+v, calls %v",
		method.Name,
		method.Exported,
		method.Pos,
		bitset.Count(method.FieldsUsed),
		method.Branches,
		method.CalledSiblings,
	)
}
