# 11. Keep the CLI a thin adapter with CI-friendly exit codes

Date: 2026-07-11

## Status

Accepted

## Context

The primary use of the tool is in CI gates and developer terminals. Exit
codes, flag parsing, signal handling, and logging are process concerns that
must not contaminate the library API (ADR 0003), yet they must be testable —
`os.Exit` in the middle of logic makes a CLI untestable.

## Decision

`cmd/modularity` is a thin adapter over the facade. `main` is two lines:
`os.Exit(run(os.Args[1:]))`. All behavior lives in `run(args []string) int`,
which parses flags into a `modularity.Config`, wires signal-based context
cancellation, calls `Analyze`, and hands the report to the reporting service
(ADR 0010). Distinct exit codes (documented in the README) separate
"analysis found violations" from "the tool itself failed", so CI can gate on
them.

## Consequences

`run` is tested directly (`main_test.go`, `main_ext_test.go`) without
spawning processes. Every capability the CLI offers must exist in the public
`Config`, which keeps the facade honest. Scripts get stable, documented exit
semantics.
