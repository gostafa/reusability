# Reusability

[![LICENSE](https://img.shields.io/github/license/gostafa/singleton)](/LICENSE) [![`Workflow for reusability Action`](https://github.com/gostafa/reusability/actions/workflows/main.yml/badge.svg)](https://github.com/gostafa/reusability/actions/workflows/main.yml) [![codecov](https://codecov.io/gh/gostafa/reusability/graph/badge.svg)](https://codecov.io/gh/gostafa/reusability) [![Go Reference](https://pkg.go.dev/badge/github.com/gostafa/reusability)](https://pkg.go.dev/github.com/gostafa/reusability)

`reusability` analyzes Go named types and reports the **reusability index** per
type. Reports show only package path, type name, and reusability — no structural
metadata (fields, methods, coupling counts, etc.).
Supporting type-level metrics (AMC, LCOM, TCC, CBO) and cyclomatic complexity
are computed internally and are not reported, selectable, or gateable on their
own.

It can run as:

* a standalone CLI;
* a plugin inside a custom `golangci-lint` binary.

Go callers import `github.com/gostafa/reusability` and use `Analyze` with
`Config` to obtain a `Report`. `ErrConfig` and `ErrAnalysis` identify failures
with `errors.Is`.

Each `Analyze` call loads fresh source. An analyzer created for `go/analysis`
or the plugin instead shares one lazy analysis result across its package
passes. It snapshots its settings at construction and caches either the
result or the failure for that instance; create a new analyzer to analyze
changed source or try again after a failed run. Shared initialization has
no fixed timeout and is attempted once.

Package loading retries typed transient system and network timeout failures
up to three total attempts. Delays start at 100 ms, double with ±20% jitter,
and have a nominal ceiling of 500 ms. A two-second scheduling budget limits
additional attempts without interrupting an ongoing load; the caller's
context cancels loading and retry waits. Untyped Go-tool errors, package
diagnostics, and unmatched patterns are not retried. Output writes, browser
launching, and profiling run once.

## Use as a CLI

### Install

```bash
go get github.com/gostafa/reusability@latest
go install github.com/gostafa/reusability/cmd/reusability@latest
```

### Run

```bash
reusability

# Check selected packages.
# reusability ./internal/...

# Write JSON or CSV.
# reusability --format=json --output=report.json ./...
# reusability --format=csv --output=report.csv ./...

# Open the HTML report.
# reusability --web ./...

# Fail when policy rules are violated.
# reusability --check \
#   --rule='**':0.6 \
#   --rule='github.com/example/project/internal/dto':0.5 \
#   ./...
```

Flags must come before package patterns:

```bash
reusability --format=json ./...
```

Useful flags:

* `--format=text|json|csv|web`
* `--output=path`
* `--tests`
* `--generated`
* `--dependency-scope=project|module|all`
* `--field-usage=direct|transitive`
* `--continue-on-error`
* `--check` with `--rule=pattern:min` (repeatable; requires `--check`)

Policy gates type-level reusability by package import path. Example:

```bash
reusability --check \
  --rule='**':0.6 \
  --rule='github.com/example/project/internal/dto':0.5 \
  ./...
```

When multiple rules match, the most specific package pattern wins: patterns
with more literal path segments take precedence, then patterns with fewer
wildcards, then longer patterns. Exact ties use the later rule. This lets a
narrow rule lower a DTO/config-heavy package's threshold without weakening the
global baseline. Rules match package import paths, not file names.

### Build from Source

```bash
git clone https://github.com/gostafa/reusability.git
cd reusability

go build -o ./bin/reusability ./cmd/reusability
./bin/reusability
```

## Use as a Golangci-Lint Plugin

The plugin must be included in a custom `golangci-lint` binary.

Create `.custom-gcl.yml`:

```yaml
version: v2.12.2
name: custom-golangci-lint
destination: ./bin
plugins:
  - module: github.com/gostafa/reusability
    import: github.com/gostafa/reusability/plugin
    path: .
```

Enable it in `.golangci.yml`:

```yaml
version: "2"

linters:
  default: all
  enable:
    - reusability

  settings:
    custom:
      reusability:
        type: module
        settings:
          dependency-scope: module
          field-usage: direct

          rules:
            - pattern: "**"
              min: 0.6
            - pattern: "github.com/example/project/internal/dto"
              min: 0.5
```

Build and run the custom linter:

```bash
golangci-lint custom -v
./bin/custom-golangci-lint run ./...
```

Always run the generated `custom-golangci-lint` binary. The standard
`golangci-lint` binary does not include the plugin.

## Exit Codes

* `0`: success
* `1`: analysis or write error
* `2`: command usage error
* `3`: policy violation
