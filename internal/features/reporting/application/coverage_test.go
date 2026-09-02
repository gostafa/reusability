package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/gostafa/reusability/internal/features/reporting/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
	"github.com/gostafa/reusability/reusability"
)

type failingSink struct {
	err error
}

func (s failingSink) Open() (io.WriteCloser, error) {
	return nil, s.err
}

type trackingWriteCloser struct {
	closed bool
	err    error
}

func (w *trackingWriteCloser) Write([]byte) (int, error) {
	return 0, w.err
}

func (w *trackingWriteCloser) Close() error {
	w.closed = true

	return nil
}

type writerSink struct {
	w io.WriteCloser
}

func (s writerSink) Open() (io.WriteCloser, error) {
	return s.w, nil
}

func TestWriteOpenAndRenderErrors(t *testing.T) {
	sentinel := errors.New("write failed")
	if err := Write(
		sampleReport(),
		domain.FormatText,
		failingSink{err: sentinel},
		domain.TextOptions{},
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("open error = %v, want sentinel", err)
	}

	w := &trackingWriteCloser{err: sentinel}
	if err := Write(
		sampleReport(),
		domain.FormatText,
		writerSink{w: w},
		domain.TextOptions{},
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("render error = %v, want sentinel", err)
	}
	if !w.closed {
		t.Fatal("writer was not closed after a render error")
	}
}

func TestJSONDebugStringsAndMarshalError(t *testing.T) {
	reportSummary := (jsonReport{
		SchemaVersion: "3",
		Tool:          jsonTool{Name: "reusability", Version: "test"},
		Packages:      []jsonPackage{{Path: "example.com/p"}},
	}).String()
	if !strings.Contains(reportSummary, "schema 3") ||
		!strings.Contains(reportSummary, "1 packages") {
		t.Fatalf("jsonReport.String() = %q", reportSummary)
	}

	packageSummary := (jsonPackage{Path: "example.com/p", Types: make([]jsonType, 1)}).String()
	if packageSummary != "example.com/p: 1 types" {
		t.Fatalf("jsonPackage.String() = %q", packageSummary)
	}

	_, err := encodeOrderedMetrics([]metrics.MetricResult{{
		Name:       metrics.MetricReusability,
		Scope:      metrics.ScopeType,
		Value:      math.NaN(),
		Applicable: true,
	}})
	if err == nil {
		t.Fatal("expected JSON encoding to reject NaN")
	}
}

func TestWriteDocsErrors(t *testing.T) {
	sentinel := errors.New("open failed")
	if err := WriteDocs(failingSink{err: sentinel}, "test"); !errors.Is(err, sentinel) {
		t.Fatalf("open error = %v, want sentinel", err)
	}

	original := docsTemplate
	docsTemplate = "missing placeholder"
	t.Cleanup(func() { docsTemplate = original })

	if err := renderDocs(io.Discard, "test"); err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	w := &trackingWriteCloser{}
	if err := WriteDocs(writerSink{w: w}, "test"); err == nil {
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
	if err := renderWeb(io.Discard, sampleReport()); err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	webTemplate = docsDataPlaceholder
	if err := renderWeb(io.Discard, sampleReport()); err == nil {
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

	// csv.Writer buffers through bufio; enough rows force a flush to the
	// underlying writer during WriteAll.
	big := sampleReport()
	pkg := big.Packages[0]
	for i := 0; i < 200; i++ {
		pkg.Types = append(pkg.Types, reusability.TypeReport{
			Name: fmt.Sprintf("T%d", i),
			Metrics: []metrics.MetricResult{{
				Name: metrics.MetricAMC, Scope: metrics.ScopeType, Value: float64(i),
				Applicable: true, Definition: "d", Reason: strings.Repeat("x", 64),
			}},
		})
	}
	big.Packages[0] = pkg

	if err := render(
		&failWriter{allow: 0, err: sentinel},
		big,
		domain.FormatCSV,
		domain.TextOptions{},
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("csv write error = %v, want sentinel", err)
	}
}

func TestJSONMarshalSeamErrors(t *testing.T) {
	original := jsonMarshal
	t.Cleanup(func() { jsonMarshal = original })

	sentinel := errors.New("marshal failed")
	jsonMarshal = func(any) ([]byte, error) { return nil, sentinel }

	if err := renderDocs(io.Discard, "test"); !errors.Is(err, sentinel) {
		t.Fatalf("renderDocs = %v, want sentinel", err)
	}
	if err := renderWeb(io.Discard, sampleReport()); !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb = %v, want sentinel", err)
	}
	if _, err := encodeOrderedMetrics([]metrics.MetricResult{{
		Name: metrics.MetricAMC, Scope: metrics.ScopeType, Value: 1, Applicable: true,
	}}); !errors.Is(err, sentinel) {
		t.Fatalf("encodeOrderedMetrics = %v, want sentinel", err)
	}
}

func TestMarshalDocsErrorViaRenderWeb(t *testing.T) {
	original := jsonMarshal
	t.Cleanup(func() { jsonMarshal = original })

	sentinel := errors.New("docs marshal failed")
	jsonMarshal = func(v any) ([]byte, error) {
		if _, ok := v.(docsPayload); ok {
			return nil, sentinel
		}

		return json.Marshal(v)
	}

	if err := renderWeb(io.Discard, sampleReport()); !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb docs error = %v, want sentinel", err)
	}
}
