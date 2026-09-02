package application

import (
	_ "embed"
	"errors"
	"io"
	"strings"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/reusability"
)

// docsTemplate is the self-contained metrics guide page: inline CSS and
// vanilla JS, native MathML formulas, no external requests.
//
//go:embed web_docs_template.html
var docsTemplate string

// docsDataPlaceholder marks where the guide JSON lands. The report template
// carries the same placeholder so its info sheets share this payload.
const docsDataPlaceholder = "__DOCS_DATA__"

// docsPayload wraps the guide entries with the tool identity for the page
// header.
type docsPayload struct {
	// Tool identifies the producing tool.
	Tool jsonTool `json:"tool"`
	// Docs are the guide entries in render order.
	Docs []jsonMetricDoc `json:"docs"`
}

// jsonMetricDoc mirrors domain.MetricDoc for the embedded payloads.
type jsonMetricDoc struct {
	// Name is the metric or column key, e.g. "amc" or "ca".
	Name string `json:"name"`
	// Label is the column heading, e.g. "AMC".
	Label string `json:"label"`
	// FullName spells the metric out.
	FullName string `json:"full_name"`
	// Scope groups the entry: type, package, or structural.
	Scope string `json:"scope"`
	// Definition is the versioned formula id; omitted for structural fields.
	Definition string `json:"definition,omitempty"`
	// FormulaMathML is display-mode <math> markup; omitted for structural
	// fields. It is the only field pages may insert as markup.
	FormulaMathML string `json:"formula_mathml,omitempty"`
	// FormulaLaTeX is the LaTeX source of record behind FormulaMathML.
	FormulaLaTeX string `json:"formula_latex,omitempty"`
	// Summary is the one-sentence meaning.
	Summary string `json:"summary"`
	// How spells out the inputs and mechanics.
	How string `json:"how"`
	// Interpretation explains when values are good or bad, and why.
	Interpretation string `json:"interpretation"`
	// NotApplicable states when the metric is n/a; omitted when always
	// applicable.
	NotApplicable string `json:"not_applicable,omitempty"`
	// Direction is "lower", "higher", or "neutral".
	Direction string `json:"direction"`
	// Bounded reports whether values live in [0, 1].
	Bounded bool `json:"bounded"`
	// Example is a small worked numeric example.
	Example string `json:"example,omitempty"`
}

// marshalDocs builds the guide JSON shared by the standalone page and the
// report page. json.Marshal HTML-escapes <, >, and &, so the MathML markup
// can never terminate its <script> element early.
func marshalDocs(toolVersion string) ([]byte, error) {
	entries := domain.MetricDocs()

	out := docsPayload{
		Tool: jsonTool{Name: reusability.ToolName, Version: toolVersion},
		Docs: make([]jsonMetricDoc, len(entries)),
	}
	for i, d := range entries {
		out.Docs[i] = jsonMetricDoc{
			Name:           d.Name,
			Label:          d.Label,
			FullName:       d.FullName,
			Scope:          string(d.Scope),
			Definition:     d.Definition,
			FormulaMathML:  d.FormulaMathML,
			FormulaLaTeX:   d.FormulaLaTeX,
			Summary:        d.Summary,
			How:            d.HowCalculated,
			Interpretation: d.Interpretation,
			NotApplicable:  d.NotApplicable,
			Direction:      d.Direction,
			Bounded:        d.Bounded,
			Example:        d.Example,
		}
	}

	return jsonMarshal(out)
}

// renderDocs writes the standalone metrics guide page: the embedded guide
// template with the docs payload injected.
func renderDocs(w io.Writer, toolVersion string) error {
	payload, err := marshalDocs(toolVersion)
	if err != nil {
		return err
	}

	if !strings.Contains(docsTemplate, docsDataPlaceholder) {
		return errors.New("docs template is missing the docs data placeholder")
	}

	page := strings.Replace(docsTemplate, docsDataPlaceholder, string(payload), 1)

	_, err = io.WriteString(w, page)

	return err
}

// WriteDocs renders the metrics guide into the sink. It needs no report:
// the page documents the tool, not one run.
func WriteDocs(sink outbound.Sink, toolVersion string) error {
	w, err := sink.Open()
	if err != nil {
		return err
	}

	renderErr := renderDocs(w, toolVersion)
	if renderErr != nil {
		_ = w.Close()

		return renderErr
	}

	return w.Close()
}
