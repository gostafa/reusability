// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

// Package main implements the layoutrefactor CLI for Go package file layout.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "print actions without writing files")
	pattern := flag.String("pattern", "./...", "go list pattern")

	flag.Parse()

	os.Exit(execute(runOptions{dryRun: *dryRun, pattern: *pattern}))
}

func execute(opts runOptions) int {
	pkgs, listErr := listPackages(opts.pattern)
	if listErr != nil {
		return exitOnListError(listErr)
	}

	return refactorAll(pkgs, opts)
}

func exitOnListError(listErr error) int {
	printErr := printTo(os.Stderr, fmt.Sprintf("list packages: %v\n", listErr))
	if printErr != nil {
		return countOne
	}

	return countOne
}

func printTo(writer *os.File, text string) error {
	written, err := fmt.Fprint(writer, text)
	if err != nil {
		return fmt.Errorf("fprint: %w", err)
	}

	if written != len(text) {
		return fmt.Errorf(fmtShortWrite, errFprintShortWrite, written, len(text))
	}

	return nil
}

func printLine(text string) error {
	written, err := fmt.Fprintln(os.Stdout, text)
	if err != nil {
		return fmt.Errorf("fprintln: %w", err)
	}

	if written < countZero {
		return fmt.Errorf("%w: %d", errPrintlnNegCount, written)
	}

	return nil
}

func refactorAll(pkgs []packageInfo, opts runOptions) int {
	failed := countZero

	for i := range pkgs {
		failed += refactorOne(&pkgs[i], opts)
	}

	if failed > countZero {
		return countOne
	}

	return countZero
}

func refactorOne(pkg *packageInfo, opts runOptions) int {
	if shouldSkipPackage(pkg) {
		return countZero
	}

	err := refactorPackage(pkg, opts)
	if err == nil {
		return countZero
	}

	err = printTo(os.Stderr, fmt.Sprintf("%s: %v\n", pkg.ImportPath, err))
	if err != nil {
		return countOne
	}

	return countOne
}
