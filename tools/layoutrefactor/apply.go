// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func commitWrites(input *commitInput) error {
	err := pickCommitErr(input)
	if err != nil {
		return fmt.Errorf(fmtCommitWrites, err)
	}

	return nil
}

func pickCommitErr(input *commitInput) error {
	runners := []func(*commitInput) error{commitDisk, printDryRun}
	idx := countZero

	if input.opts.dryRun {
		idx = countOne
	}

	err := runners[idx](input)
	if err != nil {
		return fmt.Errorf("commit runner: %w", err)
	}

	return nil
}

func printDryRun(input *commitInput) error {
	err := printLine(fmt.Sprintf("=== %s ===", input.pkg.ImportPath))
	if err != nil {
		return fmt.Errorf("print header: %w", err)
	}

	err = printWriteOps(input.writes)
	if err != nil {
		return fmt.Errorf("print writes: %w", err)
	}

	err = printDeleteOps(input.deletes)
	if err != nil {
		return fmt.Errorf("print deletes: %w", err)
	}

	return nil
}

func printWriteOps(writes []writeOp) error {
	for i := range writes {
		err := printLine(fmt.Sprintf(
			"  write %s (%d bytes)", writes[i].name, len(writes[i].content),
		))
		if err != nil {
			return fmt.Errorf("print write op: %w", err)
		}
	}

	return nil
}

func printDeleteOps(deletes []string) error {
	for i := range deletes {
		err := printLine(dryRunRemove + deletes[i])
		if err != nil {
			return fmt.Errorf("print delete op: %w", err)
		}
	}

	return nil
}

func commitDisk(input *commitInput) error {
	err := writeAll(input.pkg.Dir, input.writes)
	if err != nil {
		return fmt.Errorf("write all: %w", err)
	}

	err = deleteAll(input.pkg.Dir, input.deletes)
	if err != nil {
		return fmt.Errorf("delete all: %w", err)
	}

	return nil
}

func writeAll(dir string, writes []writeOp) error {
	for i := range writes {
		err := writeOne(dir, &writes[i])
		if err != nil {
			return fmt.Errorf("write one: %w", err)
		}
	}

	return nil
}

func writeOne(dir string, fileOp *writeOp) error {
	path := filepath.Join(dir, fileOp.name)

	err := os.WriteFile(path, fileOp.content, filePerm)
	if err != nil {
		return fmt.Errorf("write %s: %w", fileOp.name, err)
	}

	return nil
}

func deleteAll(dir string, names []string) error {
	for i := range names {
		err := deleteOne(dir, names[i])
		if err != nil {
			return fmt.Errorf("delete one: %w", err)
		}
	}

	return nil
}

func deleteOne(dir, name string) error {
	path := filepath.Join(dir, name)

	err := os.Remove(path)

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete %s: %w", name, err)
	}

	return nil
}

func planDeletes(prod, test []string, testName string) []string {
	result := appendProdDeletes(nil, prod)

	result = appendTestDeletes(result, test, testName)

	return uniqueStrings(result)
}

func appendProdDeletes(input, prod []string) []string {
	result := input

	for i := range prod {
		result = appendProdDelete(result, prod[i])
	}

	return result
}

func appendProdDelete(input []string, name string) []string {
	if name == docGoName {
		return input
	}

	return append(input, name)
}

func appendTestDeletes(input, test []string, testName string) []string {
	result := input

	for i := range test {
		result = appendTestDelete(result, test[i], testName)
	}

	return result
}

func appendTestDelete(input []string, name, testName string) []string {
	if name == testName {
		return input
	}

	return append(input, name)
}
