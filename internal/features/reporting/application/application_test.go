// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/features/reporting/ports/outbound"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

type trackingWriteCloser struct {
	closed bool
	err    error
}

func (writer *trackingWriteCloser) Write([]byte) (int, error) {
	return 0, writer.err
}

func (writer *trackingWriteCloser) Close() error {
	writer.closed = true

	return nil
}

func failingSink(err error) outbound.Sink {
	return outbound.NewSink(func() (*outbound.Stream, error) {
		return nil, err
	})
}

func writerSink(writeCloser io.WriteCloser) outbound.Sink {
	return outbound.NewSink(func() (*outbound.Stream, error) {
		return outbound.NewStream(writeCloser), nil
	})
}

func bufferSink(buf *bytes.Buffer) outbound.Sink {
	return outbound.NewSink(func() (*outbound.Stream, error) {
		return outbound.NewStream(nopCloser{buf}), nil
	})
}

func TestWriteOpenAndRenderErrors(t *testing.T) {
	sentinel := errors.New("write failed")
	if err := Write(&WriteRequest{
		Report: sampleReport(), Format: domain.FormatText,
		Sink: failingSink(sentinel), Options: domain.TextOptions{},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("open error = %v, want sentinel", err)
	}

	w := &trackingWriteCloser{err: sentinel}
	if err := Write(&WriteRequest{
		Report: sampleReport(), Format: domain.FormatText,
		Sink: writerSink(w), Options: domain.TextOptions{},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("render error = %v, want sentinel", err)
	}
	if !w.closed {
		t.Fatal("writer was not closed after a render error")
	}
}

func TestJSONDebugStringsAndMarshalError(t *testing.T) {
	reportSummary := jsonReportString(&jsonReport{
		SchemaVersion: "3",
		Tool:          jsonTool{Name: "reusability", Version: "test"},
		Packages:      []jsonPackage{{Path: "example.com/p"}},
	})
	if !strings.Contains(reportSummary, "schema 3") ||
		!strings.Contains(reportSummary, "1 packages") {
		t.Fatalf("jsonReportString() = %q", reportSummary)
	}

	packageSummary := (jsonPackage{Path: "example.com/p", Types: make([]jsonType, 1)}).String()
	if packageSummary != "example.com/p: 1 types" {
		t.Fatalf("jsonPackage.String() = %q", packageSummary)
	}

	_, err := json.Marshal(jsonMetric{
		Applicable: true,
		Value:      ptrFloat(math.NaN()),
	})
	if err == nil {
		t.Fatal("expected JSON encoding to reject NaN")
	}
}

func ptrFloat(v float64) *float64 { return &v }

func TestWriteDocsErrors(t *testing.T) {
	sentinel := errors.New("open failed")
	if err := WriteDocs(failingSink(sentinel), "test"); !errors.Is(err, sentinel) {
		t.Fatalf("open error = %v, want sentinel", err)
	}

	original := docsTemplate
	docsTemplate = "missing placeholder"
	t.Cleanup(func() { docsTemplate = original })

	if err := renderDocs(io.Discard, "test"); err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	w := &trackingWriteCloser{}
	if err := WriteDocs(writerSink(w), "test"); err == nil {
		t.Fatal("expected WriteDocs to propagate the render error")
	}
	if !w.closed {
		t.Fatal("writer was not closed after the docs render error")
	}
}

func TestRenderWebMissingPlaceholders(t *testing.T) {
	original := webTemplate
	t.Cleanup(func() { webTemplate = original })

	webTemplate = webDataPlaceholder
	if err := renderWeb(io.Discard, ptrSample()); err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	webTemplate = docsDataPlaceholder
	if err := renderWeb(io.Discard, ptrSample()); err == nil {
		t.Fatal("expected a missing report placeholder error")
	}
}

type failWriter struct {
	allow int
	err   error
	n     int
}

func (w *failWriter) Write(p []byte) (int, error) {
	w.n++
	if w.n > w.allow {
		return 0, w.err
	}

	return len(p), nil
}

func TestRenderCSVWriteErrors(t *testing.T) {
	sentinel := errors.New("csv write failed")

	big := sampleReport()
	pkg := big.Packages[0]
	for i := 0; i < 200; i++ {
		pkg.Types = append(pkg.Types, reusability.TypeReport{
			Name: fmt.Sprintf("T%d", i),
			Reusability: metrics.MetricResult{
				Name: metrics.MetricReusability, Scope: metrics.ScopeType, Value: float64(i),
				Applicable: true, Definition: "d", Reason: strings.Repeat("x", 64),
			},
		})
	}
	big.Packages[0] = pkg

	if err := render(&renderInput{
		writer: &failWriter{allow: 0, err: sentinel}, report: &big,
		format: domain.FormatCSV, opts: domain.TextOptions{},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("csv write error = %v, want sentinel", err)
	}
}

func TestJSONMarshalSeamErrors(t *testing.T) {
	sentinel := errors.New("marshal failed")
	marshal := func(any) ([]byte, error) { return nil, sentinel }

	if err := renderDocsWith(io.Discard, "test", marshal); !errors.Is(err, sentinel) {
		t.Fatalf("renderDocs = %v, want sentinel", err)
	}
	if err := renderWebWith(io.Discard, ptrSample(), marshal); !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb = %v, want sentinel", err)
	}
	if _, err := marshal(buildJSONReport(ptrSample())); !errors.Is(err, sentinel) {
		t.Fatalf("buildJSONReport marshal = %v, want sentinel", err)
	}
}

func TestMarshalDocsErrorViaRenderWeb(t *testing.T) {
	sentinel := errors.New("docs marshal failed")
	marshal := func(v any) ([]byte, error) {
		if _, ok := v.(docsPayload); ok {
			return nil, sentinel
		}

		return json.Marshal(v)
	}

	if err := renderWebWith(io.Discard, ptrSample(), marshal); !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb docs error = %v, want sentinel", err)
	}
}

func ptrSample() *reusability.Report {
	r := sampleReport()
	return &r
}

func sampleReport() reusability.Report {
	applicable := metrics.MetricResult{
		Name:       "reusability",
		Scope:      metrics.ScopeType,
		Value:      0.7,
		Applicable: true,
		Definition: "d",
	}
	na := metrics.MetricResult{
		Name:       "reusability",
		Scope:      metrics.ScopeType,
		Applicable: false,
		Reason:     "every component dropped",
		Definition: "d",
	}

	return reusability.Report{
		SchemaVersion: "6",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{{
			Path: "example.com/m/a",
			Types: []reusability.TypeReport{
				{Name: "A", Reusability: applicable},
				{Name: "B", Reusability: na},
			},
		}},
	}
}

// White-box: the JSON envelope round-trips and honors the applicability
// contract (applicable → value present; n/a → value omitted).
func TestRenderJSONContract(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := renderJSON(&buf, ptrSample())
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal([]byte(buf.String()), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkg := got["packages"].([]any)[0].(map[string]any)
	if _, ok := pkg["afferent"]; ok {
		t.Error("package afferent should not be in JSON schema v6")
	}

	typ := pkg["types"].([]any)[0].(map[string]any)
	reuse := typ["reusability"].(map[string]any)
	if reuse["applicable"] != true {
		t.Error("first reusability entry must be applicable")
	}
	if reuse["value"].(float64) != 0.7 {
		t.Errorf("reusability value = %v", reuse["value"])
	}
	if _, ok := typ["metrics"]; ok {
		t.Error("type metrics map should not appear in JSON schema v6")
	}
}

// White-box: an unknown format is rejected.
func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()

	err := render(&renderInput{
		writer: io.Discard,
		report: ptrSample(),
		format: domain.Format("xml"),
		opts:   domain.TextOptions{},
	})
	if err == nil {
		t.Fatal("unknown format should error")
	}
}

// Black-box: the metrics guide is a self-contained HTML page carrying the
// tool version, native MathML formulas, and an entry for every metric.
func TestWriteDocs(t *testing.T) {
	t.Parallel()

	captured := newCaptureSink()
	sink := captured.sink
	if err := WriteDocs(sink, "v1.2.3"); err != nil {
		t.Fatal(err)
	}

	html := captured.buf.String()
	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("guide does not start with a doctype: %.40q", html)
	}

	wanted := []string{`id="docs-data"`, `<math`, `"v1.2.3"`}
	for _, name := range metrics.ReportedMetricOrder() {
		wanted = append(wanted, `"name":"`+name+`"`)
	}

	for _, want := range wanted {
		if !strings.Contains(html, want) {
			t.Errorf("guide is missing %q", want)
		}
	}

	if strings.Contains(html, "__DOCS_DATA__") {
		t.Error("docs placeholder was not replaced")
	}

	for _, ref := range []string{`src="http`, `href="http`, `url(http`, `@import`} {
		if strings.Contains(html, ref) {
			t.Errorf("guide contains external reference %q; it must be self-contained", ref)
		}
	}
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

type captureSink struct {
	sink outbound.Sink
	buf  *bytes.Buffer
}

func newCaptureSink() captureSink {
	buf := &bytes.Buffer{}
	return captureSink{sink: bufferSink(buf), buf: buf}
}

func report() reusability.Report {
	return reusability.Report{
		SchemaVersion: "6",
		Tool:          reusability.ToolInfo{Name: "reusability", Version: "test"},
		Module:        "example.com/m",
		Packages: []reusability.PackageReport{
			{
				Path: "example.com/m/a",
				Types: []reusability.TypeReport{{
					Name: "A",
					Reusability: metrics.MetricResult{
						Name:       "reusability",
						Scope:      metrics.ScopeType,
						Value:      0.7,
						Applicable: true,
						Definition: "d",
					},
				}},
			},
		},
	}
}

// Black-box: the text format includes the module and the type row.
func TestWriteText(t *testing.T) {
	t.Parallel()

	captured := newCaptureSink()
	sink := captured.sink
	err := Write(
		&WriteRequest{
			Report:  report(),
			Format:  domain.FormatText,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	out := captured.buf.String()
	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "A") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	captured := newCaptureSink()
	sink := captured.sink
	err := Write(
		&WriteRequest{
			Report:  report(),
			Format:  domain.FormatJSON,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal(captured.buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}
	pkg := got["packages"].([]any)[0].(map[string]any)
	if _, ok := pkg["afferent"]; ok {
		t.Fatal("package afferent should not appear in JSON schema v6")
	}
	typ := pkg["types"].([]any)[0].(map[string]any)
	if typ["name"] != "A" {
		t.Fatalf("type name = %+v", typ["name"])
	}
	reuse := typ["reusability"].(map[string]any)
	if reuse["value"].(float64) != 0.7 {
		t.Fatalf("reusability = %+v", reuse)
	}
}

// Black-box: the CSV format starts with the canonical header and has a row per
// type.
func TestWriteCSV(t *testing.T) {
	t.Parallel()

	captured := newCaptureSink()
	sink := captured.sink
	if err := Write(
		&WriteRequest{
			Report:  report(),
			Format:  domain.FormatCSV,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(captured.buf).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("csv has %d rows, want header + data", len(records))
	}

	header := strings.Join(records[0], ",")
	if header != strings.Join(domain.CSVHeader(), ",") {
		t.Errorf("csv header = %q", header)
	}
}

// Black-box: the web format is a self-contained HTML page embedding the
// module and the versioned report payload, with the placeholder replaced.
func TestWriteWeb(t *testing.T) {
	t.Parallel()

	captured := newCaptureSink()
	sink := captured.sink
	err := Write(
		&WriteRequest{
			Report:  report(),
			Format:  domain.FormatWeb,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := captured.buf.String()
	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("web report does not start with a doctype: %.40q", html)
	}

	for _, want := range []string{
		`id="report-data"`,
		`"module":"example.com/m"`,
		`"schema_version":"6"`,
		`"reusability"`,
		`id="docs-data"`,
		`"formula_mathml"`,
		`id="report-table"`,
		`id="report-head"`,
		`id="report-body"`,
		`data-sort`,
		`aria-sort`,
		`class: 'help'`,
		`data-help`,
		`summarizeNode`,
		`aggregateMetricCell`,
		`meanStat`,
		`All`,
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

	if strings.Contains(html, `label: 'Position'`) || strings.Contains(html, `"kind"`) {
		t.Error("web report still defines removed structural columns")
	}

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

	captured := newCaptureSink()
	sink := captured.sink
	err := Write(
		&WriteRequest{
			Report:  rep,
			Format:  domain.FormatWeb,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := captured.buf.String()
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

	captured := newCaptureSink()
	sink := captured.sink
	err := Write(
		&WriteRequest{
			Report:  rep,
			Format:  domain.FormatWeb,
			Sink:    sink,
			Options: domain.TextOptions{},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := captured.buf.String()
	if got := strings.Count(html, `id="docs-data"`); got != 1 {
		t.Errorf("docs-data script elements = %d, want 1", got)
	}

	if !strings.Contains(html, `"name":"__DOCS_DATA__"`) {
		t.Error("hostile type name is missing from the report payload")
	}

	if !strings.Contains(html, `"formula_mathml"`) {
		t.Error("docs payload was not injected")
	}
}
