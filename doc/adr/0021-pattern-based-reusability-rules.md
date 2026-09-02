# 21. Pattern-based reusability rules

Date: 2026-09-02

## Status

Accepted

## Context

The policy system grew a large surface of structural limits (package/type/func
counts, lines, cyclomatic complexity) plus metric maps and CLI `-max`/`-min`
overrides. Package-level structural counts (vars, consts, funcs, types) were
also extracted, reported, and gateable even though the product focus is the
type-level reusability index.

ADR 0018 introduced explicit CLI thresholds so policy gates never run silently.
That intent remains, but the configuration model was harder to explain and
maintain than the value it provided.

## Decision

1. **Remove package structural counts** from extraction, the public report
   schema (bump to v5), and all renderers (text, JSON, CSV, web, docs).

2. **Replace the policy model** with `rules: [{ pattern, min }]` matched
   against full package import paths:
   - `*` matches one path segment; `**` matches zero or more segments.
   - Multiple matching rules apply the strictest (highest `min`).
   - Skip types where reusability is not applicable.

3. **Plugin**: empty/missing `rules` → `DefaultRules()` (`[{ pattern: "**", min: 0.7 }]`).

4. **CLI**: remove `-max`/`-min`; add repeatable `-rule=pattern:min`.
   `-check` requires at least one `-rule` (preserves ADR 0018's no silent gate).

5. **Keep** internal field/method/complexity/coupling extraction for reusability
   component calculation; public reports expose only type name and reusability
   (schema v6).

## Schema v6 (2026-09-02)

Public reports now carry only `package.path`, `type.name`, and
`type.reusability` (value, applicable, reason, definition). Removed from
all output formats: Ca/Ce, kind, exported flag, position, field/method
counts and details, and the metrics map. Internal computation is unchanged;
`--explain` still surfaces n/a reasons and dropped-component notes.

## Consequences

- Simpler configuration and violation messages focused on reusability.
- Structural parts of ADR 0017/0018 policy config are superseded for gating.
- Schema v5 breaks consumers of package func/var/const counts in JSON reports.
- Schema v6 breaks consumers of Ca/Ce, type structural facts, and the metrics
  map; each type now has a single `reusability` object.
- golangci-lint settings must migrate from `package`/`type`/`funcs`/`metrics`
  blocks to `rules`.
