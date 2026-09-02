package version

import "testing"

// White-box: the embedded version string is always populated.
func TestVersionIsSet(t *testing.T) {
	t.Parallel()

	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}
