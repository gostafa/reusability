# 9. Produce deterministic schema-versioned reports

Date: 2026-07-11

## Status

Accepted

## Context

Reports are consumed by scripts and diffed in CI. Two runs over the same
source must produce byte-identical output, or diffs become noise —
especially since computation is parallel (ADR 0008) and map iteration in Go
is randomized. Consumers also need to detect when the report format itself
changes.

## Decision

Every `Report` carries a `SchemaVersion` constant alongside tool name and
version. Packages, types, and metrics are emitted in a stable, sorted order
(`internal/shared/metrics/order.go` defines the canonical metric ordering),
and `Analyze` documents determinism as part of its contract
(`modularity.go`).

## Consequences

Golden-file tests and CI diffing work reliably, and downstream tooling can
gate on `SchemaVersion` instead of sniffing fields. The discipline cuts both
ways: any change to the report shape or ordering is a schema change and must
bump the version, and every new metric must register in the canonical order.
