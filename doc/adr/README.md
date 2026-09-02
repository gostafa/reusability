# Architecture Decision Records

Decision log for modularity, managed with `task adrs:*`
(see ADR 0014). Regenerate this list with `task adrs:generate -- toc`.

* [1. Record architecture decisions](0001-record-architecture-decisions.md)
* [2. Adopt hexagonal architecture with feature slices](0002-adopt-hexagonal-architecture-with-feature-slices.md)
* [3. Expose a single public facade package](0003-expose-a-single-public-facade-package.md)
* [4. Enforce domain purity with an architecture guard test](0004-enforce-domain-purity-with-an-architecture-guard-test.md)
* [5. Ban the legacy project identifier](0005-ban-the-legacy-project-identifier.md)
* [6. Keep dependencies minimal](0006-keep-dependencies-minimal.md)
* [7. Load packages via golang.org/x/tools/go/packages](0007-load-packages-via-golang-org-x-tools-go-packages.md)
* [8. Compute metrics concurrently with a shared worker pool](0008-compute-metrics-concurrently-with-a-shared-worker-pool.md)
* [9. Produce deterministic schema-versioned reports](0009-produce-deterministic-schema-versioned-reports.md)
* [10. Support text, JSON, and CSV output through sink adapters](0010-support-text-json-and-csv-output-through-sink-adapters.md)
* [11. Keep the CLI a thin adapter with CI-friendly exit codes](0011-keep-the-cli-a-thin-adapter-with-ci-friendly-exit-codes.md)
* [12. Use Task with vendored taskfiles synced by taskotter](0012-use-task-with-vendored-taskfiles-synced-by-taskotter.md)
* [13. Run lint and test coverage in GitHub Actions CI](0013-run-lint-and-test-coverage-in-github-actions-ci.md)
* [14. Use adrs to manage architecture decision records](0014-use-adrs-to-manage-architecture-decision-records.md)
* [15. Render a self-contained HTML web report](0015-render-a-self-contained-html-web-report.md)
* [16. Generate a self-contained web help page for metrics](0016-generate-a-self-contained-web-help-page-for-metrics.md)
* [17. Superseded: Add go.yaml.in/yaml/v4 for policy config](0017-add-yaml-v4-for-policy-config.md)
* [18. Enforce a modularity policy gate](0018-enforce-a-modularity-policy-gate.md)
* [19. Dual-mode: CLI library and golangci-lint Module Plugin](0019-dual-mode-cli-and-golangci-module-plugin.md)
* [20. Split into a single-public-metric reusability linter](0020-split-into-a-single-public-metric-reusability-linter.md)
