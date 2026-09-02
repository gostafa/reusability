// Package domain contains the pure modularity policy model.
//
// It defines conditions (limits) over a report's structural facts and metrics,
// the recommended default policy, and the evaluation of a report into ordered
// violations. It performs no I/O; adapters build a Policy from CLI flags,
// golangci-lint settings, or any future configuration source.
package domain
