# 17. Superseded: Add go.yaml.in/yaml/v4 for policy config

Date: 2026-07-13

## Status

Superseded

Superseded on 2026-07-17 when the standalone CLI moved to explicit
`-max`/`-min` threshold flags only. The YAML policy adapter and
`go.yaml.in/yaml/v4` dependency were removed; the golangci-lint plugin remains
inline-settings-only.

## Current Implementation

The standalone CLI builds policies only from repeated `-max` and `-min` flags.
The golangci-lint plugin builds policies only from inline `.golangci.yml`
settings. Neither entry point reads `.modularity.yml`.

## Context

Historically, the standalone CLI policy gate (ADR 0018) could read its
conditions from a `.modularity.yml` file. YAML is what developers expect for
tool configuration
— it sits beside `.golangci.yml` and the project's other config — but the
standard library has no YAML decoder. ADR 0006 keeps the module to a single
direct dependency and requires a strong justification, and ideally its own
ADR, before adding another.

Hand-rolling a YAML subset was considered and rejected. This file is consumed
in other people's CI, where a parser that silently mis-reads valid YAML — the
kind of edge case a subset misses — is worse than an extra dependency: it would
turn a passing gate into a false green. Correct decoding, unknown-key
detection, and precise line numbers in errors are exactly what a mature library
already provides.

## Decision

This ADR originally added `go.yaml.in/yaml/v4` as the second direct dependency.
This is the canonical home of the Go YAML library (the successor to
`gopkg.in/yaml.v3`); v4 is chosen over the legacy import path so the project
tracks the maintained line. Its use is quarantined to one infrastructure adapter,
`internal/infrastructure/policyconfig`, which maps the decoded document onto the
pure policy domain. The domain, the facade, and every other package remain
yaml-free; the architecture guard (ADR 0004) keeps `policy/domain` pure, so the
dependency cannot leak inward.

The golangci-lint Module Plugin added by ADR 0019 does not use this adapter. Its
policy is decoded directly from the custom linter's inline `.golangci.yml`
settings, and it never loads or discovers `.modularity.yml`.

## Consequences

This decision was valid while the CLI accepted policy files, but it no longer
describes the current implementation. The policy domain stayed pure, which made
the removal small: the CLI now constructs policies from flags, the plugin
constructs policies from `.golangci.yml` inline settings, and no YAML dependency
is needed.
