// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpenWithControlledOpener(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if err := Open("report.html"); err == nil {
		t.Fatal("missing opener accepted")
	}
	name, _ := OpenCommand(runtime.GOOS, "")
	data, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := Open("-test.run=^$"); err != nil {
		t.Fatal(err)
	}
	cmd := command(os.Args[0], []string{"-test.run=^$"})
	if err := startCommand(cmd); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

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
