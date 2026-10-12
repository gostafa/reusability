// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"os"
	"testing"
)

const (
	testZero    = 0
	testOne     = 1
	testSeven   = 7
	flagVersion = "--version"
)

func TestMainDelegatesToCLI(t *testing.T) {
	var (
		gotArgs []string
		gotCode int
	)

	runtime := mainRuntime{
		run: func(args []string) int {
			gotArgs = append([]string(nil), args...)

			return testSeven
		},
		exit: func(code int) { gotCode = code },
	}

	runtime.start([]string{flagVersion})

	if len(gotArgs) != testOne || gotArgs[testZero] != flagVersion {
		t.Fatalf("args = %v", gotArgs)
	}

	if gotCode != testSeven {
		t.Fatalf("exit code = %d, want %d", gotCode, testSeven)
	}
}

func TestMainRunsCLI(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	defer func() {
		if got := recover(); got != "unexpected call to os.Exit(0) during test" {
			t.Fatalf("exit = %v", got)
		}
	}()
	os.Args = []string{"reusability", flagVersion}
	main()
}
