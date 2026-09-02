package browser_test

import (
	"testing"

	"github.com/gostafa/reusability/internal/infrastructure/browser"
)

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
		name, args := browser.OpenCommand(tt.goos, "r.html")
		if name != tt.want {
			t.Errorf("OpenCommand(%q) name = %q, want %q", tt.goos, name, tt.want)
		}

		if len(args) == 0 || args[len(args)-1] != "r.html" {
			t.Errorf("OpenCommand(%q) args = %v, want the path last", tt.goos, args)
		}
	}
}
