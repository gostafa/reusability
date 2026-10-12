// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package zerovalue_test

import (
	"testing"

	"github.com/gostafa/reusability/internal/shared/zerovalue"
)

func TestOf(t *testing.T) {
	t.Parallel()
	if zerovalue.Of[int]() != 0 || zerovalue.Of[[]string]() != nil {
		t.Fatal("zero values must be the language-defined defaults")
	}
}
