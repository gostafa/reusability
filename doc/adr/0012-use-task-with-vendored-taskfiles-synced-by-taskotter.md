# 12. Use Task with vendored taskfiles synced by taskotter

Date: 2026-07-11

## Status

Accepted

## Context

Build, lint, test, and tooling commands need to run identically on macOS,
Linux, Windows, and CI. Ad-hoc shell scripts drift per platform, and
tooling recipes shared across repositories drift per repository unless
something keeps them in sync.

## Decision

[Task](https://taskfile.dev) is the task runner. The root `Taskfile.yml`
defines project tasks (build, test, coverage) and includes vendored,
reusable modules under `taskfiles/` — `go`, `adrs`, `cargo`, `staticcheck`,
`actionlint`, `yamllint`, `zizmor`, `uv` — each exposed as a namespace
(e.g. `task adrs:new`). The taskotter bot keeps those vendored modules in
sync with their upstream via automated pull requests
(`.github/workflows/taskotter-sync.yml`; see the `chore(taskotter): sync
taskfiles` merges in history).

## Consequences

Developers and CI (ADR 0013) run the same entry points, and tool
installation is self-serve (`task <tool>:install`). Upstream fixes arrive
as reviewable PRs. Vendored files under `taskfiles/` must not be edited by
hand — local changes would be overwritten by the next sync — so
project-specific tasks belong in the root `Taskfile.yml`.
