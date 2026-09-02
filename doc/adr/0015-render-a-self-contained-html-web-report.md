# 15. Render a self-contained HTML web report

Date: 2026-07-11

## Status

Accepted

## Context

Terminal tables and CSV rows are poor at exploring a large report: users
want to sort by a metric, filter to a package, and see values at a glance.
Shipping a web app would violate the minimal-dependency rule (ADR 0006) and
complicate distribution; requiring a network or a server would make reports
unusable offline and unarchivable.

## Decision

`web` is a fourth report format inside the reporting slice (ADR 0010). The
renderer embeds a single HTML template (`go:embed`) and injects the report
as a JSON payload into a `<script type="application/json">` element; the
payload wraps — never alters — the versioned JSON schema (ADR 0009), adding
only the module path. The page is fully self-contained: inline CSS and
vanilla JavaScript, no external requests, usable from `file://`. Its chrome
is monochrome with a black/white theme toggle; green/orange/red appear only
on metric values, reusing the terminal renderer's quality thresholds. It
offers a collapsible directory tree using the text report's path hierarchy
plus flat types and packages views with column sorting,
text/package/applicability filters, and motion-reduced animations. Directory
rows carry subtree means; package rows retain exact package metrics and carry
subtree type means. `json.Marshal`'s HTML escaping keeps hostile identifiers
from terminating the script element.

The CLI gains `--web` as a shorthand for `-format=web`; without `-output`
the report defaults to `modularity-report.html` and opens in the platform
browser via a new infrastructure adapter — only when stdout is a terminal,
so CI stays quiet (ADR 0011).

## Consequences

Reports are shareable as one file and render identically offline and in CI
artifacts. The vanilla-JS constraint means no build step and no npm surface,
at the cost of hand-rolled table logic. Schema consumers are unaffected: the
web payload is additive packaging around the versioned JSON schema. Opening a
browser is best-effort; failures degrade to a logged warning, never a failed
run.
