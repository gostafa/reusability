// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func listPackages(pattern string) ([]packageInfo, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, goBinaryName, "list", "-json", "-e")

	cmd.Args = append(cmd.Args, pattern)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	pkgs, err := decodePackages(out)
	if err != nil {
		return nil, fmt.Errorf("decode packages: %w", err)
	}

	return pkgs, nil
}

func decodePackages(out []byte) ([]packageInfo, error) {
	var pkgs []packageInfo

	dec := json.NewDecoder(bytes.NewReader(out))

	for dec.More() {
		pkg, err := decodeOnePackage(dec)
		if err != nil {
			return nil, fmt.Errorf("decode one package: %w", err)
		}

		pkgs = appendPackage(pkgs, &pkg)
	}

	return pkgs, nil
}

func decodeOnePackage(dec *json.Decoder) (packageInfo, error) {
	var raw map[string]any

	err := dec.Decode(&raw)
	if err != nil {
		return packageInfo{}, fmt.Errorf("json decode: %w", err)
	}

	return packageFromRaw(raw), nil
}

func packageFromRaw(raw map[string]any) packageInfo {
	return packageInfo{
		ImportPath: anyString(raw["ImportPath"]),
		Dir:        anyString(raw["Dir"]),
		Name:       anyString(raw["Name"]),
	}
}

func anyString(value any) string {
	text, ok := value.(string)

	if !ok {
		return emptyString
	}

	return text
}

func appendPackage(pkgs []packageInfo, pkg *packageInfo) []packageInfo {
	if pkg.Dir == emptyString {
		return pkgs
	}

	return append(pkgs, *pkg)
}

func shouldSkipPackage(pkg *packageInfo) bool {
	if strings.Contains(pkg.ImportPath, testdataSeg) {
		return true
	}

	return strings.HasSuffix(pkg.Dir, selfToolPath)
}
