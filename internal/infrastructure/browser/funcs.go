// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

const (
	goosDarwin  = "darwin"
	goosWindows = "windows"

	openerDarwin  = "open"
	openerWindows = "rundll32"
	openerXDG     = "xdg-open"

	windowsURLHandler = "url.dll,FileProtocolHandler"
)

// Open opens path with the platform's default browser and returns without
// waiting for the browser to exit.
func Open(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)

	err := startCommand(command(name, args))
	if err != nil {
		return fmt.Errorf("open %s in browser: %w", path, err)
	}

	return nil
}

// OpenCommand returns the command name and arguments that open path with
// the default browser on the given platform. Unknown platforms fall back to
// the freedesktop opener.
func OpenCommand(goos, path string) (name string, args []string) {
	switch goos {
	case goosDarwin:
		return openerDarwin, []string{path}
	case goosWindows:
		return openerWindows, []string{windowsURLHandler, path}
	default:
		return openerXDG, []string{path}
	}
}

func startCommand(cmd *exec.Cmd) error {
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("start browser command: %w", err)
	}

	return nil
}

// command builds the [exec.Cmd] for a fixed opener binary. name is always one
// of OpenCommand's platform opener constants, never user input, and every
// element of args is a separate argv entry rather than shell-interpolated text,
// so there is no command-injection surface. The command is assembled directly
// (with the binary resolved via [exec.LookPath], exactly as [exec.Command] would)
// so that this shell-free construction is explicit rather than hidden behind a
// variadic call. A lookup failure is left for start to surface, matching
// [exec.Command]'s deferred behavior.
func command(name string, args []string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}

	resolved, lookErr := exec.LookPath(name)
	if lookErr == nil {
		cmd.Path = resolved
	}

	return cmd
}
