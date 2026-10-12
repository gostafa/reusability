// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import "testing"

// Black-box: each platform maps to its documented opener with the path as
// the final argument; unknown platforms fall back to the freedesktop opener.
func TestOpenCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		goos string
		want string
	}{
		{"darwin", "open"},
		{"windows", "rundll32"},
		{"linux", "xdg-open"},
		{"freebsd", "xdg-open"},
	}
	for _, tt := range tests {
		name, args := OpenCommand(tt.goos, "r.html")
		if name != tt.want {
			t.Errorf("OpenCommand(%q) name = %q, want %q", tt.goos, name, tt.want)
		}

		if len(args) == 0 || args[len(args)-1] != "r.html" {
			t.Errorf("OpenCommand(%q) args = %v, want the path last", tt.goos, args)
		}
	}
}
