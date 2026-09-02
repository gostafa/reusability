# 6. Keep dependencies minimal

Date: 2026-07-11

## Status

Accepted

## Context

A static-analysis tool gets imported into other people's builds and run in
their CI. Every dependency it drags along adds supply-chain surface, version
conflicts, and install weight. Most of what the tool needs — parsing flags,
encoding JSON/CSV, walking files, concurrency — the Go standard library
already provides.

## Decision

`go.mod` carries exactly one direct dependency: `golang.org/x/tools`, needed
for type-checked package loading (ADR 0007). Its transitive requirements
(`golang.org/x/mod`, `golang.org/x/sync`) are the only indirect entries.
Everything else — CLI, output encoding, worker pool, bitset — is built on
the standard library. Adding a new direct dependency requires a strong
justification and, ideally, its own ADR.

## Consequences

The module is cheap to audit, vendor, and embed, and `go install` stays
fast. In exchange, small utilities are hand-rolled in `internal/shared/`
(`workerpool`, `bitset`) instead of imported, and those must be tested to
the same standard as any library.
