// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package version

import (
	"runtime/debug"
	"testing"
)

// Black-box: consumers read a non-empty version string.
func TestVersionExported(t *testing.T) {
	t.Parallel()

	if Version() == "" {
		t.Fatal("version.Version() must not be empty")
	}
}

func TestVersionFallbackAndRelease(t *testing.T) {
	t.Parallel()
	if got := readVersion(func() (*debug.BuildInfo, bool) { return nil, false }); got != "dev" {
		t.Fatalf("unavailable version = %q", got)
	}
	for _, value := range []string{"", "(devel)", "v1.2.3"} {
		want := value
		if value != "v1.2.3" {
			want = "dev"
		}
		if got := versionFromInfo(value); got != want {
			t.Fatalf("version(%q) = %q", value, got)
		}
	}
}
