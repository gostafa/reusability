# 16. Generate a self-contained web help page for metrics

Date: 2026-07-11

## Status

Accepted

## Context

The semantics of the reported metrics — formulas, inputs, good/bad
directions, n/a conventions — live only in README prose, undiscoverable
from the tool itself. `--help` prints one line per flag and nothing about
what an LCOM of 0.8 means. Real formulas need typeset math, but the
self-containment rule for web output (ADR 0015) and the minimal-dependency
rule (ADR 0006) forbid CDN-loaded or vendored math libraries.

## Decision

`--help --web` (either order) writes a self-contained metrics guide page to
the OS temp directory (`modularity-help-*.html`), logs the path, and
opens it in the platform browser — only when stdout is a terminal, so CI
stays quiet (ADR 0011). The file is always written and the run exits 0;
plain `--help` keeps its usage output and exit code 2. Because `--help`
aborts flag parsing before `-web` is seen, the CLI detects the combination
by handling `flag.ErrHelp` and scanning the raw arguments for a truthy
`-web` token before any `--` terminator.

One catalog — `MetricDocs()` in the reporting domain, pure data guarded by
the domain-purity test (ADR 0004) — is the single source for every entry:
label, formula, meaning, mechanics, interpretation, n/a conditions, and a
worked example, with direction and boundedness pinned to the renderers'
quality map by test. Formulas are authored as MathML Core, which browsers
typeset natively without any library; the LaTeX source of record travels as
code comments and `alttext`, so no KaTeX/MathJax is needed.

The catalog is marshaled once and injected as a JSON payload into two
embedded templates: the standalone guide page and the existing report page,
which gains a `?` affordance on every documented column header opening an
anchored info sheet with the same content. The trusted docs payload is
injected before the untrusted report payload so hostile identifiers cannot
spoof the docs placeholder; the guide is not a fifth report format — there
is no report to format — so it ships as `WriteDocs` beside `Write`.

## Consequences

Metric semantics ship inside the binary and render offline as one file,
with real typeset math and zero new dependencies. The report page explains
its own columns without duplicating a word of prose. MathML requires a
2023+ browser; older ones degrade to readable symbol runs annotated with
the LaTeX `alttext`. Help output grows a second artifact: `--help --web`
exits 0 where plain `--help` stays a usage error, and the temp-dir file is
disposable by design.
