# 20. Split into a single-public-metric reusability linter

Date: 2026-09-02

## Status

Accepted

## Context

`github.com/gostafa/modularity` computed eight metrics across two natural
scopes: type-level (`amc`, `lcom`, `tcc`, `cbo`, `reusability`) and
package-level (`abstractness`, `instability`, `distance`). Callers could
select any subset. The metric dependency graph already drew the line:
`reusability → {lcom, amc, cbo}`.

The combined tool asked users to choose among internals they should not have
to know. The type-level linter's public contract is one number: how reusable
a named type is.

This repository is a full copy of that tree, not a shared core. Zero coupling
between the two linters is the point. ADRs 0001–0019 are inherited history
from the combined tool.

## Decision

Ship `github.com/gostafa/reusability` as a standalone linter whose only
reported, selectable, and gateable metric is `reusability`.

* AMC, LCOM, TCC, CBO, and cyclomatic complexity remain in the compute
  closure and may appear in docs prose as inputs to the formula. They are
  not columns, flags, or policy keys. A config naming them is rejected as
  an unknown policy metric.
* The CLI has no `--metrics` flag. The pipeline display set is hardcoded to
  `{reusability}`.
* Default policy gates `reusability >= 0.7` plus the structural limits that
  still apply. Package-level metrics are gone.
* The facade package, binary, and golangci plugin are all named `reusability`.

## Consequences

Reports, JSON/CSV/HTML, and the golangci plugin all expose one metric column.
Hidden inputs stay testable as internals. The combined `modularity` module is
unchanged by this split.
