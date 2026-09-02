# 18. Enforce a modularity policy gate

Date: 2026-07-13

## Status

Accepted

## Context

The tool computed metrics but could not act on them: every successful run
exited `0`, so keeping a codebase within a budget meant post-processing JSON in
each project's CI. ADR 0011 anticipated this — it reserved a distinct exit code
to separate "analysis found violations" from "the tool itself failed" — but the
violations path was never built.

## Decision

Add a policy gate. A `Policy` is a set of conditions — `max`/`min` bounds on
every structural fact (types, exported/unexported funcs, coupling, fields,
methods) and every metric. Metric bounds may be scoped to package metrics or
type metrics so a package metric such as `distance` is not confused with a type
metric such as `reusability`; legacy flat metric maps remain accepted as global
bounds. `Evaluate(report, policy)` is a pure domain function returning ordered
violations; it skips conditions on metrics that are not applicable to an entity
so n/a cells never fail a build. Numeric comparisons use a relative tolerance
of `1e-12 × max(1, |value|, |threshold|)`, preventing representation noise at
an exact boundary from producing a violation while preserving meaningful
threshold crossings.

For the CLI, conditions come only from repeatable `-max`/`-min` threshold flags.
The flags imply `-check`; passing `-check` without any thresholds is a usage
error so a CI gate cannot succeed with an empty policy. The golangci-lint Module
Plugin (ADR 0019) instead decodes `package`, `type`, and legacy/global
`metrics` sections directly from `.golangci.yml`. Omitting every inline policy
section selects the recommended defaults; providing any section creates a
self-contained inline policy without merging omitted limits from those defaults.

The gate is opt-in and explicit: it runs only when `-check` or a threshold flag
is given.
Following ADR 0011, the CLI writes the report to stdout, prints the violation
summary to stderr, and exits `3`. Policy configuration errors are usage errors
(`2`).

When the Module Plugin is enabled, policy enforcement is part of the linter run
and violations are emitted as golangci-lint diagnostics rather than CLI exit
code `3` summaries.

## Consequences

The tool is now a CI gate with stable, documented exit semantics. The policy
model is a pure feature slice (`internal/features/policy/domain`) wired by the
thin CLI, exactly as reporting is (ADR 0003), so the facade stays the minimal
`Analyze → Report` library. Gated metrics are auto-added to the display set,
which means enabling a condition can add a column to the report. CLI users must
spell their budget in flags, while plugin users may choose either explicit
inline limits or the recommended defaults. The tolerance is deliberately much
smaller than displayed metric precision, so it fixes boundary noise without
weakening configured budgets.
