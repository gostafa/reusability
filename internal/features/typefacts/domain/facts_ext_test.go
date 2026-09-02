package domain_test

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
)

// Black-box: keys are deterministic and distinguish types by name.
func TestTypeKeyContract(t *testing.T) {
	t.Parallel()

	a := typefacts.TypeKey("p", "A")
	if a != typefacts.TypeKey("p", "A") {
		t.Fatal("TypeKey must be deterministic")
	}

	if a == typefacts.TypeKey("p", "B") {
		t.Fatal("distinct names must produce distinct keys")
	}

	if a == typefacts.TypeKey("q", "A") {
		t.Fatal("distinct packages must produce distinct keys")
	}
}

// Black-box: the kind constants are mutually distinct.
func TestKindConstantsDistinct(t *testing.T) {
	t.Parallel()

	kinds := map[typefacts.TypeKind]bool{
		typefacts.KindStruct:    true,
		typefacts.KindInterface: true,
		typefacts.KindOther:     true,
	}
	if len(kinds) != 3 {
		t.Fatalf("type kinds are not distinct: %v", kinds)
	}
}
