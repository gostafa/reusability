# 2. Adopt hexagonal architecture with feature slices

Date: 2026-07-11

## Status

Accepted

## Context

The tool computes many modularity metrics (cohesion, complexity, coupling,
reusability, …) from Go source. Metric logic is long-lived and mathematical,
while the machinery around it — compiler frontends, CLI flags, encoders,
filesystems — churns. Mixing the two makes the metric code hard to test and
hard to trust.

## Decision

Structure `internal/` as ports-and-adapters with vertical feature slices:

- `internal/features/<feature>/{domain,application,ports}` for each analysis
  concern (architecture, cohesion, complexity, projectanalysis, reporting,
  reusability, typefacts). Domain holds pure logic, application orchestrates,
  ports declare inbound/outbound interfaces.
- `internal/infrastructure/` holds the adapters that implement the ports:
  `analyzer` (pipeline wiring), `goloader` (compiler frontend), `profiling`,
  and `sinks` (output writers).
- `internal/shared/` is the shared kernel: `metrics`, `bitset`, `version`,
  `workerpool`.

## Consequences

Domain packages are testable without a compiler or filesystem, and adapters
are replaceable behind ports. The cost is more packages and some indirection.
The boundary is enforced mechanically by a guard test (see ADR 0004), so the
layout does not depend on reviewer vigilance.
