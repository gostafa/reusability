# 5. Ban the legacy project identifier

Date: 2026-07-11

## Status

Accepted

## Context

The project was renamed to modularity from an earlier name (the word
"go" joined with the word "metrics"). Leftover references to the old name in
imports, docs, or identifiers cause confusion and can silently reintroduce
the old module path.

## Decision

The rename is enforced mechanically. `TestForbiddenIdentifier` in
`guard_test.go` walks the entire repository and fails if any file contains
the legacy identifier, case-insensitively. The needle is assembled at
runtime from two string halves so the test file cannot violate the rule it
enforces — and every other file, including this ADR, must refer to the old
name only indirectly.

## Consequences

The rename stays total: no stale import paths, binary names, or doc
references can land, because the test fails in CI. All prose in the
repository (READMEs, ADRs, comments) must describe the old name without
spelling it, which reads slightly awkwardly but keeps the guarantee absolute.
