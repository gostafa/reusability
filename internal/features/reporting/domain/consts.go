// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// DocScopeType marks a type-level metric entry.
	DocScopeType DocScope = "type"
	// DocScopePackage marks a package-level metric entry.
	DocScopePackage DocScope = "package"
	// DocScopeStructural marks a counted column (Ca, Funcs, Fields, …).
	DocScopeStructural DocScope = "structural"

	// DirectionLower means lower metric values are better.
	DirectionLower = "lower"
	// DirectionHigher means higher metric values are better.
	DirectionHigher = "higher"
	// DirectionNeutral means the metric has no preferred direction.
	DirectionNeutral = "neutral"

	mathMLReusabilityRI = `<math display="block" alttext="RI = w_c C + w_k (1 - K) + w_t T + w_d D">` +
		`<mrow><mi>RI</mi><mo>=</mo><msub><mi>w</mi><mi>c</mi></msub><mi>C</mi><mo>+</mo>` +
		`<msub><mi>w</mi><mi>k</mi></msub><mo stretchy="false">(</mo><mn>1</mn><mo>−</mo>` +
		`<mi>K</mi><mo stretchy="false">)</mo><mo>+</mo><msub><mi>w</mi><mi>t</mi></msub>` +
		`<mi>T</mi><mo>+</mo><msub><mi>w</mi><mi>d</mi></msub><mi>D</mi></mrow></math>`

	mathMLReusabilityC = `<math display="block" alttext="C = 1 - LCOM">` +
		`<mrow><mi>C</mi><mo>=</mo><mn>1</mn><mo>−</mo><mi>LCOM</mi></mrow></math>`

	mathMLReusabilityK = `<math display="block" alttext="K = \\frac{CBO}{CBO + 1}">` +
		`<mrow><mi>K</mi><mo>=</mo><mfrac><mi>CBO</mi>` +
		`<mrow><mi>CBO</mi><mo>+</mo><mn>1</mn></mrow></mfrac></mrow></math>`

	mathMLReusabilityT = `<math display="block" alttext="T = \\frac{1}{1 + \\max(0, AMC - 1)}">` +
		`<mrow><mi>T</mi><mo>=</mo><mfrac><mn>1</mn><mrow><mn>1</mn><mo>+</mo>` +
		`<mi>max</mi><mo stretchy="false">(</mo><mn>0</mn><mo>,</mo><mi>AMC</mi>` +
		`<mo>−</mo><mn>1</mn><mo stretchy="false">)</mo></mrow></mfrac></mrow></math>`

	mathMLReusabilityD = `<math display="block" ` +
		`alttext="D = \\frac{documented exported members}{exported members}">` +
		`<mrow><mi>D</mi><mo>=</mo><mfrac>` +
		`<mtext>documented exported members</mtext>` +
		`<mtext>exported members</mtext></mfrac></mrow></math>`

	formulaLaTeXReusability = `RI = w_c C + w_k (1 - K) + w_t T + w_d D` + newline +
		`C = 1 - LCOM` + newline +
		`K = \\frac{CBO}{CBO + 1}` + newline +
		`T = \\frac{1}{1 + \\max(0,\\ AMC - 1)}` + newline +
		`D = \\frac{\\text{documented exported members}}{\\text{exported members}}`

	howCalculatedReusability = "Four normalized 0–1 components combine with weights " +
		"w_c = 0.35 (cohesion C, from LCOM), w_k = 0.25 (coupling K, from CBO, " +
		"contributing 1 − K), w_t = 0.25 (testability T, from average method " +
		"complexity), and w_d = 0.15 (documentation D) by default. CLI weight " +
		"flags or golangci-lint reusability-weights settings can override them. " +
		"A component whose input is not applicable is dropped and the remaining " +
		"weights are renormalized to sum to 1, keeping the index in 0–1; dropped " +
		"components are listed in the metric's reason. LCOM, AMC, TCC, and CBO " +
		"are computed internally and are not reported, selectable, or gateable " +
		"on their own."

	interpretationReusability = "A high index combines cohesive methods, few " +
		"collaborators, simple control flow, and a documented exported surface — " +
		"the properties that make a type safe to lift out and reuse elsewhere. " +
		"It is experimental: treat it as a triage hint and read the reason field " +
		"when a component was dropped."

	summaryReusability = "An experimental composite of cohesion, coupling, " +
		"testability, and documentation."

	notApplicableReusability = "Only when every weighted component is dropped — " +
		"for example a type with no methods, no fields, and no exported members."

	exampleReusability = "C = 0.8, 1 − K = 0.75, T = 0.5, D = 1.0 with default " +
		"weights: RI = 0.35·0.8 + 0.25·0.75 + 0.25·0.5 + 0.15·1.0 ≈ 0.74."

	// FormatText renders a human-readable report.
	FormatText Format = "text"
	// FormatJSON renders the versioned JSON schema.
	FormatJSON Format = "json"
	// FormatCSV renders one row per entity and metric.
	FormatCSV Format = "csv"
	// FormatWeb renders a self-contained interactive HTML report.
	FormatWeb Format = "web"

	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"

	naCell = "–"

	emptyString = ""
	newline     = "\n"
	pathDot     = "."

	floatBits = 64

	indexZero = 0
	countOne  = 1
	countTwo  = 2

	qualityHigh   = 0.66
	qualityMedium = 0.33

	teeBranch    = "├── "
	cornerBranch = "└── "
	teePad       = "│   "
	cornerPad    = "    "

	spaceString    = " "
	pathTypeHeader = "PATH / TYPE"
	notesLabel     = "notes"

	biasHigherBetter scoreBias = 4
	biasLowerBetter  scoreBias = 5

	errWrapWriteRow = "write row: %w"
)
