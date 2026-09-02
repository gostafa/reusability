package application_test

import (
	"bytes"
	"strings"
	"testing"

	reporting "github.com/gostafa/reusability/internal/features/reporting/application"
	"github.com/gostafa/reusability/internal/features/reporting/domain"
)

// Black-box: the web format is a self-contained HTML page embedding the
// module and the versioned report payload, with the placeholder replaced.
func TestWriteWeb(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatWeb, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	html := sink.buf.String()
	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("web report does not start with a doctype: %.40q", html)
	}

	for _, want := range []string{
		`id="report-data"`,
		`"module":"example.com/m"`,
		`"schema_version":"1"`,
		`"functions":[{"name":"Run"`,
		`"variables":[{"name":"Config"`,
		`"method_details":[{"name":"Do"`,
		`"reusability"`,
		`id="docs-data"`,
		`"formula_mathml"`,
		`id="report-table"`,
		`id="report-head"`,
		`id="report-body"`,
		`id="kind-pills"`,
		`id="export-pills"`,
		`data-sort`,
		`aria-sort`,
		`class: 'help'`,
		`data-help`,
		`Cyclomatic complexity for the function`,
		`summarizeNode`,
		`aggregateMetricCell`,
		`meanStat`,
		`Packages`,
		`Vars`,
		`Consts`,
		`Funcs`,
		`All`,
		`Exported`,
		`Types`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("web report is missing %q", want)
		}
	}

	for _, placeholder := range []string{"__REPORT_DATA__", "__DOCS_DATA__"} {
		if strings.Contains(html, placeholder) {
			t.Errorf("placeholder %s was not replaced", placeholder)
		}
	}

	if strings.Contains(html, `label: 'Position'`) {
		t.Error("web report still defines a Position column")
	}

	// Self-containment: nothing on the page may fetch an external resource.
	for _, ref := range []string{`src="http`, `href="http`, `url(http`, `@import`} {
		if strings.Contains(html, ref) {
			t.Errorf("web report contains external reference %q; it must be self-contained", ref)
		}
	}
}

// Black-box: hostile identifiers cannot terminate the payload's script
// element early — json.Marshal HTML-escapes every angle bracket.
func TestWriteWebEscapesScriptTerminator(t *testing.T) {
	t.Parallel()

	rep := report()
	rep.Packages[0].Types[0].Name = "</script><script>alert(1)</script>"

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(rep, domain.FormatWeb, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	html := sink.buf.String()
	if strings.Contains(html, "</script><script>alert(1)") {
		t.Error("payload contains an unescaped script terminator")
	}

	if !strings.Contains(html, `</script>`) {
		t.Error("angle brackets in the payload are not escaped")
	}
}

// Black-box: a hostile identifier spelling the docs placeholder cannot
// hijack the docs script element — the trusted docs payload is injected
// before the untrusted report payload.
func TestWriteWebPayloadCannotSpoofDocsPlaceholder(t *testing.T) {
	t.Parallel()

	rep := report()
	rep.Packages[0].Types[0].Name = "__DOCS_DATA__"

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(rep, domain.FormatWeb, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	html := sink.buf.String()
	if got := strings.Count(html, `id="docs-data"`); got != 1 {
		t.Errorf("docs-data script elements = %d, want 1", got)
	}

	// The hostile name survives as literal report data...
	if !strings.Contains(html, `"name":"__DOCS_DATA__"`) {
		t.Error("hostile type name is missing from the report payload")
	}

	// ...while the docs script still carries the genuine guide payload.
	if !strings.Contains(html, `"formula_mathml"`) {
		t.Error("docs payload was not injected")
	}
}
