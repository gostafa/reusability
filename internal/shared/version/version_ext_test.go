package version_test

import (
	"testing"

	"github.com/gostafa/reusability/internal/shared/version"
)

// Black-box: consumers read a non-empty version string.
func TestVersionExported(t *testing.T) {
	t.Parallel()

	if version.Version == "" {
		t.Fatal("version.Version must not be empty")
	}
}
