# 19. Dual-mode: CLI library and golangci-lint Module Plugin

Date: 2026-07-17

## Status

Accepted

## Context

`modularity` already ships as a standalone CLI and library facade
(`modularity.Analyze`, ADR 0003) with a policy gate (ADR 0018). Teams that
standardize on golangci-lint want the same policy enforcement inside their
existing lint binary, without giving up the independent CLI for reports and
CI exit code `3`.

golangci-lint's recommended extension path is the [Module Plugin
System](https://golangci-lint.run/docs/plugins/module-plugins/), which loads
plugins that implement `register.LinterPlugin` from
`github.com/golangci/plugin-module-register` and return `go/analysis`
analyzers. A naive per-package Pass rewrite cannot compute coupling metrics
(Ca/Ce, abstractness, instability, distance): those need a whole-module load.

ADR 0006 keeps dependencies minimal and requires an ADR for new direct deps.

## Decision

Keep the hexagonal analysis pipeline unchanged. Add an adapter layer:

1. Public package `analyzer` — builds a `go/analysis.Analyzer` that runs
   `modularity.Analyze` once (`sync.Once`) over configured patterns,
   evaluates the modularity policy, and emits diagnostics for the current
   package's violations. Positions come from the Pass AST (no report schema
   change). This package does **not** import plugin-module-register.
2. Public package `plugin` — registers `modularity` via
   `register.Plugin`, decodes custom settings into `analyzer.Settings`, and
   returns `LoadModeTypesInfo`.
3. Add `github.com/golangci/plugin-module-register` as a direct dependency,
   quarantined to `plugin/`.

Plugin policy thresholds are decoded inline from the custom linter's
`.golangci.yml` settings. The plugin does not load or discover
`.modularity.yml`; neither mode reads a policy file. The standalone CLI takes
policy thresholds from repeatable `-max`/`-min` flags, while the plugin takes
them from inline settings. Omitting all inline policy sections selects the
recommended defaults. Providing any of the `package`, `type`, or legacy/global
`metrics` sections defines the complete plugin policy; omitted limits are not
merged back from the defaults. The inline shape matches the policy domain's
`min`/`max` limits, including numeric shorthand for `max`, but does not carry a
file `version` key.

The plugin also accepts analysis settings that map to the public
`modularity.Config` facade, including package patterns, tests/generated
inclusion, dependency scope, field usage mode, workers, build tags,
continue-on-error, and `reusability-weights`. Reusability weight fields are
optional and merge with the default component weights so an explicit `0` can
disable one component without resetting omitted components.

Settings decoding is strict. The former `config` file-path setting, unknown
structural or limit keys, and metrics placed in the wrong scope are errors
instead of silent no-ops.

Consumers build a custom binary with `.custom-gcl.yml` pointing at this
module's `plugin` import, then enable `modularity` under
`linters.settings.custom`.

## Consequences

One analysis core serves both modes. The Module Plugin dependency is present
in `go.mod` but unused unless a consumer imports `plugin`. Coupling metrics
remain correct because the adapter still performs a whole-module analyze.
Plugin configuration is self-contained in `.golangci.yml`, while CLI users can
use explicit `-max`/`-min` threshold flags and CLI analysis flags.

Amend ADR 0003 to allow the two lint packages alongside the metrics facade.
