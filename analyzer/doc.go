// Package analyzer provides a go/analysis Analyzer that enforces a
// reusability policy. It reuses the root reusability facade for whole-module
// metrics, then reports policy violations as diagnostics for the package
// under analysis.
//
// This package has no dependency on golangci-lint. Use the sibling plugin
// package to register the analyzer as a Module Plugin.
package analyzer
