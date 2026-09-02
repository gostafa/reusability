package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// startCommand launches a prepared command without waiting for it to exit.
// Tests replace it to observe the launch without opening a real browser.
var startCommand = func(cmd *exec.Cmd) error { return cmd.Start() }

// Open opens path with the platform's default browser and returns without
// waiting for the browser to exit.
func Open(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)
	if err := startCommand(command(name, args)); err != nil {
		return fmt.Errorf("open %s in browser: %w", path, err)
	}

	return nil
}

// command builds the *exec.Cmd for a fixed opener binary. name is always one
// of OpenCommand's platform opener constants, never user input, and every
// element of args is a separate argv entry rather than shell-interpolated text,
// so there is no command-injection surface. The command is assembled directly
// (with the binary resolved via exec.LookPath, exactly as exec.Command would)
// so that this shell-free construction is explicit rather than hidden behind a
// variadic call. A lookup failure is left for startCommand to surface, matching
// exec.Command's deferred behaviour.
func command(name string, args []string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}
	if resolved, err := exec.LookPath(name); err == nil {
		cmd.Path = resolved
	}

	return cmd
}

// OpenCommand returns the command name and arguments that open path with
// the default browser on the given platform. Unknown platforms fall back to
// the freedesktop opener.
func OpenCommand(goos, path string) (name string, args []string) {
	switch goos {
	case "darwin":
		return "open", []string{path}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default:
		return "xdg-open", []string{path}
	}
}
