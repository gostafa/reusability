# 8. Compute metrics concurrently with a shared worker pool

Date: 2026-07-11

## Status

Accepted

## Context

Metric computation is CPU-bound and independent per package: once the facts
are extracted (ADR 0007), each package's metrics can be computed in
isolation. On multi-core machines a sequential pass leaves most of the
hardware idle on large repositories.

## Decision

A small, dependency-free worker pool lives in `internal/shared/workerpool`
and is used by the analysis pipeline to fan out per-package metric
computation. The degree of parallelism is part of the public API:
`Config.Workers`, defaulting sensibly when unset and forwarded through
`inbound.Options`. Workers respect context cancellation so Ctrl-C aborts
promptly.

## Consequences

Analysis scales with available cores without any external dependency.
Because results complete out of order, the pipeline must reassemble them
into a stable order before reporting — that requirement is codified in
ADR 0009. The pool is shared infrastructure, so any future parallel stage
reuses it instead of spawning ad-hoc goroutines.
