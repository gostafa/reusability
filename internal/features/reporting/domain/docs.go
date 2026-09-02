package domain

import "github.com/gostafa/reusability/internal/shared/metrics"

// DocScope groups metrics-guide entries by the kind of entity they
// describe: type metrics, package metrics, or the structural columns that
// are counted rather than computed.
type DocScope string

const (
	// DocScopeType marks a type-level metric entry.
	DocScopeType DocScope = "type"
	// DocScopePackage marks a package-level metric entry.
	DocScopePackage DocScope = "package"
	// DocScopeStructural marks a counted column (Ca, Funcs, Fields, …).
	DocScopeStructural DocScope = "structural"
)

// Doc directions: whether smaller or larger values are better, or no
// universal direction exists. They mirror qualityByMetric.
const (
	DirectionLower   = "lower"
	DirectionHigher  = "higher"
	DirectionNeutral = "neutral"
)

// MetricDoc explains one reported metric or structural field to a human:
// what it means, how it is computed, and how to judge its values. It is
// the single source behind the standalone metrics guide (--help --web) and
// the report page's per-column explanations.
type MetricDoc struct {
	// Name is the metric or column key, e.g. "amc" or "ca".
	Name string
	// Label is the column heading, e.g. "AMC".
	Label string
	// FullName spells the metric out, e.g. "Average Method Complexity".
	FullName string
	// Scope groups the entry: type metric, package metric, or structural.
	Scope DocScope
	// Definition is the versioned formula id; empty for structural fields.
	Definition string
	// FormulaMathML holds display-mode <math> markup. MathML Core only, so
	// browsers typeset it natively and the page stays self-contained.
	// Empty for structural fields.
	FormulaMathML string
	// FormulaLaTeX is the LaTeX source of record behind FormulaMathML.
	FormulaLaTeX string
	// Summary is the one-sentence meaning.
	Summary string
	// HowCalculated spells out the inputs and mechanics.
	HowCalculated string
	// Interpretation explains when values are good or bad, and why.
	Interpretation string
	// NotApplicable states when the metric is n/a; empty means always
	// applicable.
	NotApplicable string
	// Direction is "lower", "higher", or "neutral", matching the quality
	// coloring of the renderers.
	Direction string
	// Bounded reports whether values live in [0, 1].
	Bounded bool
	// Example is a small worked numeric example.
	Example string
}

const formulaReusability = `<math display="block" alttext="RI = w_c C + w_k (1 - K) + w_t T + w_d D"><mrow><mi>RI</mi><mo>=</mo><msub><mi>w</mi><mi>c</mi></msub><mi>C</mi><mo>+</mo><msub><mi>w</mi><mi>k</mi></msub><mo stretchy="false">(</mo><mn>1</mn><mo>−</mo><mi>K</mi><mo stretchy="false">)</mo><mo>+</mo><msub><mi>w</mi><mi>t</mi></msub><mi>T</mi><mo>+</mo><msub><mi>w</mi><mi>d</mi></msub><mi>D</mi></mrow></math>
<math display="block" alttext="C = 1 - LCOM"><mrow><mi>C</mi><mo>=</mo><mn>1</mn><mo>−</mo><mi>LCOM</mi></mrow></math>
<math display="block" alttext="K = \frac{CBO}{CBO + 1}"><mrow><mi>K</mi><mo>=</mo><mfrac><mi>CBO</mi><mrow><mi>CBO</mi><mo>+</mo><mn>1</mn></mrow></mfrac></mrow></math>
<math display="block" alttext="T = \frac{1}{1 + \max(0, AMC - 1)}"><mrow><mi>T</mi><mo>=</mo><mfrac><mn>1</mn><mrow><mn>1</mn><mo>+</mo><mi>max</mi><mo stretchy="false">(</mo><mn>0</mn><mo>,</mo><mi>AMC</mi><mo>−</mo><mn>1</mn><mo stretchy="false">)</mo></mrow></mfrac></mrow></math>
<math display="block" alttext="D = \frac{documented exported members}{exported members}"><mrow><mi>D</mi><mo>=</mo><mfrac><mtext>documented exported members</mtext><mtext>exported members</mtext></mfrac></mrow></math>`

// MetricDocs returns the guide entries for the reported metric. Internal
// inputs (AMC, LCOM, TCC, CBO) are described in the reusability formula;
// they are not documented as selectable metrics.
func MetricDocs() []MetricDoc {
	return []MetricDoc{
		{
			Name:           metrics.MetricReusability,
			Label:          abbrev(metrics.MetricReusability),
			FullName:       "Experimental Reusability Index",
			Scope:          DocScopeType,
			Definition:     metrics.DefinitionReusability,
			FormulaMathML:  formulaReusability,
			FormulaLaTeX:   `RI = w_c C + w_k (1 - K) + w_t T + w_d D` + "\n" + `C = 1 - LCOM` + "\n" + `K = \frac{CBO}{CBO + 1}` + "\n" + `T = \frac{1}{1 + \max(0,\ AMC - 1)}` + "\n" + `D = \frac{\text{documented exported members}}{\text{exported members}}`,
			Summary:        "An experimental composite of cohesion, coupling, testability, and documentation.",
			HowCalculated:  "Four normalized 0–1 components combine with weights w_c = 0.35 (cohesion C, from LCOM), w_k = 0.25 (coupling K, from CBO, contributing 1 − K), w_t = 0.25 (testability T, from average method complexity), and w_d = 0.15 (documentation D) by default. CLI weight flags or golangci-lint reusability-weights settings can override them. A component whose input is not applicable is dropped and the remaining weights are renormalized to sum to 1, keeping the index in 0–1; dropped components are listed in the metric's reason. LCOM, AMC, TCC, and CBO are computed internally and are not reported, selectable, or gateable on their own.",
			Interpretation: "A high index combines cohesive methods, few collaborators, simple control flow, and a documented exported surface — the properties that make a type safe to lift out and reuse elsewhere. It is experimental: treat it as a triage hint and read the reason field when a component was dropped.",
			NotApplicable:  "Only when every weighted component is dropped — for example a type with no methods, no fields, and no exported members.",
			Direction:      DirectionHigher,
			Bounded:        true,
			Example:        "C = 0.8, 1 − K = 0.75, T = 0.5, D = 1.0 with default weights: RI = 0.35·0.8 + 0.25·0.75 + 0.25·0.5 + 0.15·1.0 ≈ 0.74.",
		},
	}
}
