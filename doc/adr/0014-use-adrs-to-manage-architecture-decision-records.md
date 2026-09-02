# 14. Use adrs to manage architecture decision records

Date: 2026-07-11

## Status

Accepted

## Context

The decisions in this log previously lived only in code, guard tests, and
contributors' heads. ADR 0001 commits the project to recording decisions;
that needs a concrete tool and workflow so records stay uniform and cheap
to create.

## Decision

Use [adrs](https://github.com/joshrotenberg/adrs), a Rust,
adr-tools-compatible command-line manager, driven through the vendored
Taskfile module (ADR 0012): `task adrs:init`, `task adrs:new -- "Title"`,
`task adrs:list`, `task adrs:generate -- toc`. Records live in `doc/adr`
(the tool default) as plain numbered Markdown in the Nygard format.

## Consequences

Creating a new record is one command, numbering and templates are handled
by the tool, and the records are plain Markdown reviewable in PRs. The
`adrs` binary installs via cargo, but only ADR authors need it — readers
just need the Markdown files. New architectural decisions should land with
an ADR in the same pull request.
