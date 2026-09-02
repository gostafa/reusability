package inbound

import (
	"strings"
	"testing"
)

// White-box: the debug Stringer summarizes a package result.
func TestPackageResultString(t *testing.T) {
	t.Parallel()

	pr := PackageResult{
		Path:  "example.com/m/p",
		Types: make([]TypeResult, 1),
	}

	s := pr.String()
	for _, want := range []string{"example.com/m/p", "1 types"} {
		if !strings.Contains(s, want) {
			t.Errorf("String()=%q missing %q", s, want)
		}
	}
}
