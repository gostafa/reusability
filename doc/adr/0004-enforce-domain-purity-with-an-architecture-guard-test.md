# 4. Enforce domain purity with an architecture guard test

Date: 2026-07-11

## Status

Accepted

## Context

The hexagonal layout of ADR 0002 only works if domain packages actually stay
pure. Architecture rules that live in documentation decay: one convenient
import of `go/types` or `os` in a domain package and the boundary is gone,
usually unnoticed in review.

## Decision

`TestDomainPurity` in `guard_test.go` parses the imports (via
`go/parser.ParseFile` with `ImportsOnly`) of `internal/shared/metrics` and of
every `internal/features/*/domain` package, and fails if any of them imports
the compiler stack (`go/*`, `golang.org/x/tools`), encoding (`encoding/json`,
`encoding/csv`), the filesystem (`os`, `io/fs`, `io/ioutil`,
`path/filepath`), `flag`, or `log`.

## Consequences

A boundary violation is a test failure, caught locally and in CI (ADR 0013)
rather than in review. New feature domains are picked up automatically
because the test discovers `internal/features/*/domain` directories at
runtime. The forbidden-prefix list is a maintained artifact: new impure
dependency classes must be added to it explicitly.
