// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"fmt"
	"go/ast"
	"strings"
)

func formatPackageDoc(doc *ast.CommentGroup) string {
	if doc == nil {
		return emptyString
	}

	return strings.Join(docLines(doc), newline)
}

func docLines(doc *ast.CommentGroup) []string {
	var lines []string

	for i := range doc.List {
		lines = appendDocLine(lines, doc.List[i].Text)
	}

	return lines
}

func appendDocLine(lines []string, raw string) []string {
	text := strings.TrimSpace(strings.TrimPrefix(raw, "//"))

	if text == emptyString {
		return lines
	}

	return append(lines, text)
}

func defaultPackageDoc(name string) string {
	return fmt.Sprintf("Package %s provides reusable analysis utilities.", name)
}

func buildDocGo(license, pkgName, doc string) ([]byte, error) {
	var builder strings.Builder

	err := writeDocSections(&builder, &docSectionsInput{
		license: license,
		pkgName: pkgName,
		doc:     doc,
	})
	if err != nil {
		return nil, fmt.Errorf("doc sections: %w", err)
	}

	return []byte(builder.String()), nil
}

func writeDocSections(builder *strings.Builder, input *docSectionsInput) error {
	err := writeDocHeader(builder, input.license)
	if err != nil {
		return fmt.Errorf("doc header: %w", err)
	}

	err = writeDocComments(builder, input.doc)
	if err != nil {
		return fmt.Errorf("doc comments: %w", err)
	}

	err = writeDocPackage(builder, input.pkgName)
	if err != nil {
		return fmt.Errorf("doc package: %w", err)
	}

	return nil
}

func writeDocHeader(builder *strings.Builder, license string) error {
	err := appendBuilder(builder, license)
	if err != nil {
		return fmt.Errorf("license: %w", err)
	}

	err = appendBuilder(builder, doubleNewline)
	if err != nil {
		return fmt.Errorf("header newline: %w", err)
	}

	return nil
}

func writeDocComments(builder *strings.Builder, doc string) error {
	for line := range strings.SplitSeq(doc, newline) {
		err := appendBuilderParts(builder, "// ", line, newline)
		if err != nil {
			return fmt.Errorf("doc comment line: %w", err)
		}
	}

	return nil
}

func writeDocPackage(builder *strings.Builder, pkgName string) error {
	err := appendBuilderParts(builder, "package ", pkgName, newline)
	if err != nil {
		return fmt.Errorf("package line: %w", err)
	}

	return nil
}

func appendBuilderParts(builder *strings.Builder, parts ...string) error {
	for i := range parts {
		err := appendBuilder(builder, parts[i])
		if err != nil {
			return fmt.Errorf("builder part: %w", err)
		}
	}

	return nil
}

func appendBuilder(builder *strings.Builder, text string) error {
	written, err := builder.WriteString(text)
	if err != nil {
		return fmt.Errorf("builder write: %w", err)
	}

	if written != len(text) {
		return fmt.Errorf(fmtShortWrite, errBuilderShortWrite, written, len(text))
	}

	return nil
}
