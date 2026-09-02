# 3. Expose a single public facade package

Date: 2026-07-11

## Status

Accepted (amended 2026-07-17)

## Context

The analyzer is consumed as a CLI, as an importable library, and — after
ADR 0019 — as a golangci-lint Module Plugin. Library consumers need a stable,
documented metrics API, but the implementation — pipelines, loaders, metric
packages — needs the freedom to change shape without breaking anyone. Lint
integration needs a small public surface that golangci-lint can blank-import;
that surface must not collapse into the metrics facade.

## Decision

The root package `modularity` remains the public metrics API. It exposes
`Analyze(ctx context.Context, config Config) (Report, error)` plus the
`Config` and `Report` types (`modularity.go`, `config.go`, `report.go`).

Two additional public packages exist solely for lint integration (ADR 0019):

- `analyzer` — a `go/analysis` Analyzer with no golangci-lint dependency
- `plugin` — the Module Plugin registration entry point

Every other implementation package lives under `internal/`, which the Go
compiler makes unimportable from outside the module.

## Consequences

Internals can be refactored freely — the ports, features, and adapters of
ADR 0002 are invisible to consumers. The metrics facade stays the stable
library entry point. Consumers who only want reports never import `plugin`
or `analyzer`. The flip side is deliberate: nothing inside `internal/` can
be reused by third parties, and any capability worth exposing must be
surfaced through the facade or the lint packages.
