// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"
	"testing"
)

// White-box: the canonical cross-package key.
func TestTypeKey(t *testing.T) {
	t.Parallel()

	if got := TypeKey("example.com/m/pkg", "Widget"); got != "example.com/m/pkg.Widget" {
		t.Fatalf("TypeKey = %q", got)
	}
}

// White-box: the debug Stringers stay informative and panic-free.
func TestStringers(t *testing.T) {
	t.Parallel()

	pf := &ProjectFacts{
		ModulePath: "m",
		Packages:   make([]PackageFacts, 2),
		Types:      make([]TypeFacts, 3),
	}
	if s := projectFactsString(
		pf,
	); !strings.Contains(s, "2 packages") ||
		!strings.Contains(s, "3 types") {
		t.Errorf("projectFactsString = %q", s)
	}

	tf := &TypeFacts{Name: "W", Kind: KindInterface}
	if s := typeFactsString(tf); !strings.Contains(s, `"W"`) {
		t.Errorf("typeFactsString = %q", s)
	}
	if got := KindInterface.String(); got != "1" {
		t.Errorf("KindInterface.String = %q, want decimal discriminant", got)
	}

	te := &TypeExtract{Name: "E"}
	if s := typeExtractString(te); !strings.Contains(s, `"E"`) {
		t.Errorf("typeExtractString = %q", s)
	}
}

// Black-box: keys are deterministic and distinguish types by name.
func TestTypeKeyContract(t *testing.T) {
	t.Parallel()

	a := TypeKey("p", "A")
	if a != TypeKey("p", "A") {
		t.Fatal("TypeKey must be deterministic")
	}

	if a == TypeKey("p", "B") {
		t.Fatal("distinct names must produce distinct keys")
	}

	if a == TypeKey("q", "A") {
		t.Fatal("distinct packages must produce distinct keys")
	}
}

// Black-box: the kind constants are mutually distinct.
func TestKindConstantsDistinct(t *testing.T) {
	t.Parallel()

	kinds := map[TypeKind]bool{KindStruct: true, KindInterface: true, KindOther: true}
	if len(kinds) != 3 {
		t.Fatalf("type kinds are not distinct: %v", kinds)
	}
}
