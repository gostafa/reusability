// Package reusability analyzes Go modules and reports a type-level
// reusability index.
//
// Call Analyze with a Config to load packages, compute the index (and the
// cohesion, complexity, and coupling inputs it needs internally), and
// receive a deterministic Report. Config selects patterns, field-usage
// mode, dependency scope, and component weights. The supporting metrics
// are never reported, selectable, or gateable on their own.
//
// For policy enforcement via go/analysis, use the sibling analyzer package.
// To register that analyzer as a golangci-lint Module Plugin, blank-import
// the sibling plugin package.
package reusability
