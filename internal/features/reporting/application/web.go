package application

import (
	_ "embed"
	"errors"
	"io"
	"strings"

	"github.com/gostafa/reusability/reusability"
)

// webTemplate is the self-contained HTML page: inline CSS and vanilla JS,
// no external requests, usable straight from file://.
//
//go:embed web_template.html
var webTemplate string

// webDataPlaceholder marks where the JSON payload lands in the template.
const webDataPlaceholder = "__REPORT_DATA__"

// webPayload wraps the versioned JSON report with the module path for the
// page header. The v1 JSON schema itself stays untouched.
type webPayload struct {
	// Module is the analyzed main module's path, when known.
	Module string `json:"module"`
	// Report is the same document the JSON format emits.
	Report jsonReport `json:"report"`
}

// renderWeb writes the interactive HTML report: the embedded template with
// the JSON payloads injected. json.Marshal HTML-escapes <, >, and &, so the
// payloads can never terminate their <script> elements early. The trusted
// compile-time docs payload is injected before the report payload, whose
// untrusted identifiers could otherwise spoof the docs placeholder.
func renderWeb(w io.Writer, report reusability.Report) error {
	payload, err := jsonMarshal(webPayload{
		Module: report.Module,
		Report: buildJSONReport(report),
	})
	if err != nil {
		return err
	}

	docs, err := marshalDocs(report.Tool.Version)
	if err != nil {
		return err
	}

	if !strings.Contains(webTemplate, docsDataPlaceholder) {
		return errors.New("web template is missing the docs data placeholder")
	}

	page := strings.Replace(webTemplate, docsDataPlaceholder, string(docs), 1)

	if !strings.Contains(page, webDataPlaceholder) {
		return errors.New("web template is missing the report data placeholder")
	}

	page = strings.Replace(page, webDataPlaceholder, string(payload), 1)

	_, err = io.WriteString(w, page)

	return err
}
