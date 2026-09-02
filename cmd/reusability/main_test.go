package main

import (
	"os"
	"testing"
)

func TestMainDelegatesToCLI(t *testing.T) {
	originalArgs, originalRun, originalExit := os.Args, run, exit
	t.Cleanup(func() {
		os.Args, run, exit = originalArgs, originalRun, originalExit
	})

	os.Args = []string{"reusability", "--version"}
	var gotArgs []string
	var gotCode int
	run = func(args []string) int {
		gotArgs = append([]string(nil), args...)

		return 7
	}
	exit = func(code int) { gotCode = code }

	main()

	if len(gotArgs) != 1 || gotArgs[0] != "--version" {
		t.Fatalf("args = %v", gotArgs)
	}
	if gotCode != 7 {
		t.Fatalf("exit code = %d, want 7", gotCode)
	}
}
