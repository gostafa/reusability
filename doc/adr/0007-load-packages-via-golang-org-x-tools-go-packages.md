# 7. Load packages via golang.org/x/tools/go/packages

Date: 2026-07-11

## Status

Accepted

## Context

The metrics need fully type-checked information: method sets, field usage,
embedding, generics, build-tag and test-file variants, and module-aware
package resolution. Reimplementing that on top of `go/build` and a manual
type-check loop is error-prone and would break with every toolchain change.

## Decision

Package loading is delegated to `golang.org/x/tools/go/packages`, wrapped in
the `internal/infrastructure/goloader` adapter (`loader.go`). The adapter
loads the configured patterns (honoring build tags, test inclusion, and
context cancellation) and `extract.go` distills the type-checked results
into the pure fact types of the `typefacts` feature. Per ADR 0004, the
compiler stack is only importable here and in the other infrastructure
adapters — never in domain code.

## Consequences

Loading is module-aware and tracks toolchain evolution through routine
`x/tools` upgrades (the module's single direct dependency, ADR 0006).
Domain code consumes plain fact structs and never sees an AST, so a
different frontend could be swapped in behind the port. Package loading
dominates the tool's runtime, which is why computation downstream is
parallelized (ADR 0008).
