// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"fmt"
)

// String summarizes the package result for debugging.
func (p PackageResult) String() string {
	return fmt.Sprintf("%s: %d types", p.Path, len(p.Types))
}
