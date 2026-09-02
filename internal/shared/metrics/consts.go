// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

const (
	// MetricAMC is the average method complexity metric name.
	MetricAMC = "amc"
	// DefinitionAMC identifies the AMC formula version.
	DefinitionAMC = "reusability/amc-v1"
	// MetricCBO is the coupling-between-objects metric name.
	MetricCBO = "cbo"
	// DefinitionCBO identifies the CBO formula version.
	DefinitionCBO = "reusability/cbo-v1"
	// MetricLCOM is the lack-of-cohesion metric name.
	MetricLCOM = "lcom"
	// DefinitionLCOM identifies the LCOM formula version.
	DefinitionLCOM = "reusability/lcom-v1"
	// ScopeType marks a metric computed per named type.
	ScopeType MetricScope = "type"
	// ScopePackage marks a metric computed once per package.
	ScopePackage MetricScope = "package"
	// MetricReusability is the experimental reusability index metric name.
	MetricReusability = "reusability"
	// DefinitionReusability identifies the reusability formula version.
	DefinitionReusability = "reusability/reusability-v1"
	// ComponentCohesion names the cohesion reusability component.
	ComponentCohesion = "cohesion"
	// ComponentCoupling names the coupling reusability component.
	ComponentCoupling = "coupling"
	// ComponentTestability names the testability reusability component.
	ComponentTestability = "testability"
	// ComponentDocumentation names the documentation reusability component.
	ComponentDocumentation = "documentation"
	// MetricTCC is the tight class cohesion metric name.
	MetricTCC = "tcc"
	// DefinitionTCC identifies the TCC formula version.
	DefinitionTCC = "reusability/tcc-v1"

	zero                  = 0
	one                   = 1
	two                   = 2
	defaultWeightCohesion = 0.30
	defaultWeightCoupling = 0.35
	defaultWeightTest    = 0.20
	defaultWeightDocs     = 0.15

	reasonNoMethods = "type has no methods"
	listSep         = ", "
)
