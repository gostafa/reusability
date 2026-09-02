// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/shared/bitset"
)

func TestMethodFactsString(t *testing.T) {
	fieldsUsed := bitset.NewFieldSet(3)
	fieldsUsed.Set(0)
	fieldsUsed.Set(2)

	facts := &MethodFacts{
		Name:           "Save",
		Exported:       true,
		Pos:            Position{File: "store.go", Line: 12, Column: 3},
		FieldsUsed:     fieldsUsed,
		Branches:       BranchStats{Ifs: 1, LogicalOps: 2},
		CalledSiblings: []int{1, 3},
	}

	got := facts.String()
	for _, want := range []string{
		`method "Save"`,
		"exported true",
		"uses 2 fields",
		"calls [1 3]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}
