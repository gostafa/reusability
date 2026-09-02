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
	if s := pf.String(); !strings.Contains(s, "2 packages") || !strings.Contains(s, "3 types") {
		t.Errorf("ProjectFacts.String = %q", s)
	}

	tf := &TypeFacts{Name: "W", Kind: KindInterface}
	if s := tf.String(); !strings.Contains(s, `"W"`) {
		t.Errorf("TypeFacts.String = %q", s)
	}

	mf := &MethodFacts{Name: "Do"}
	if s := mf.String(); !strings.Contains(s, `"Do"`) {
		t.Errorf("MethodFacts.String = %q", s)
	}

	te := &TypeExtract{Name: "E"}
	if s := te.String(); !strings.Contains(s, `"E"`) {
		t.Errorf("TypeExtract.String = %q", s)
	}
}
