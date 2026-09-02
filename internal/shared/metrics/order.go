package metrics

// TypeMetricOrder is the internal compute order of type-level metrics.
// AMC, LCOM, TCC, and CBO feed reusability and are never reported.
func TypeMetricOrder() []string {
	return []string{
		MetricAMC, MetricLCOM, MetricTCC,
		MetricCBO, MetricReusability,
	}
}

// PackageMetricOrder is the fixed rendering order of package-level metrics.
// This linter reports no package-level metric, so the order is empty.
func PackageMetricOrder() []string {
	return nil
}

// ReportedMetricOrder is the single public metric this linter renders.
func ReportedMetricOrder() []string {
	return []string{MetricReusability}
}

// dependencies is the metric-level dependency graph. Selecting a metric
// pulls its dependencies into the compute set; internal inputs such as the
// method-field matrix are handled inside features and need no entry.
var dependencies = map[string][]string{
	MetricReusability: {MetricLCOM, MetricAMC, MetricCBO},
}

// Closure expands a selected display set to the full compute set: the
// transitive closure over metric dependencies, deduplicated, in a
// deterministic order. A metric computed only to satisfy a dependency is
// not rendered unless also selected.
func Closure(selected []string) []string {
	seen := make(map[string]bool, len(selected))

	var visit func(name string)

	visit = func(name string) {
		if seen[name] {
			return
		}

		seen[name] = true
		for _, dep := range dependencies[name] {
			visit(dep)
		}
	}
	for _, name := range selected {
		visit(name)
	}

	closure := make([]string, 0, len(seen))
	for _, name := range TypeMetricOrder() {
		if seen[name] {
			closure = append(closure, name)
		}
	}

	for _, name := range PackageMetricOrder() {
		if seen[name] {
			closure = append(closure, name)
		}
	}

	return closure
}
