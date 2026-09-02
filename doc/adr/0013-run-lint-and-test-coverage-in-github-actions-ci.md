# 13. Run lint and test coverage in GitHub Actions CI

Date: 2026-07-11

## Status

Accepted

## Context

The architecture guarantees of this project live in tests (ADRs 0004, 0005)
and in linters; they only hold if something runs them on every change.
Quality checks that exist but are not enforced centrally get skipped.

## Decision

GitHub Actions is the CI platform. `.github/workflows/main.yml` runs on
every push and pull request with dedicated jobs for linting and for tests
with coverage, driving the same Task targets developers use locally
(ADR 0012). Linting covers Go (`staticcheck`), YAML (`yamllint`), workflow
definitions (`actionlint`), and workflow security (`zizmor`). A separate
`taskotter-sync.yml` workflow handles taskfile synchronization.

## Consequences

Local and CI behavior stay identical because both go through Task. The
guard tests make architecture violations a red build, not a review comment.
Workflow files themselves are linted and security-scanned, which the
`zizmor` job exists to enforce.
