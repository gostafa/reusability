# reusability

`reusability` analyzes Go named types and reports the **reusability index** per
type. Reports show only package path, type name, and reusability — no structural
metadata (fields, methods, coupling counts, etc.).
Supporting type-level metrics (AMC, LCOM, TCC, CBO) and cyclomatic complexity
are computed internally and are not reported, selectable, or gateable on their
own.

It can run as:

* a standalone CLI;
* a plugin inside a custom `golangci-lint` binary.

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

### Build from source

```bash
git clone https://github.com/gostafa/reusability.git
cd reusability

go build -o ./bin/reusability ./cmd/reusability
./bin/reusability
```

## Use as a golangci-lint plugin

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

## Exit codes

* `0`: success
* `1`: analysis or write error
* `2`: command usage error
* `3`: policy violation
