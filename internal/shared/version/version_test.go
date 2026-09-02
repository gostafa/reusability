// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package version

import "testing"

// Black-box: consumers read a non-empty version string.
func TestVersionExported(t *testing.T) {
	t.Parallel()

	if Version() == "" {
		t.Fatal("version.Version() must not be empty")
	}
}
