package application_test

import (
	"bytes"
	"strings"
	"testing"

	reporting "github.com/gostafa/reusability/internal/features/reporting/application"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Black-box: the metrics guide is a self-contained HTML page carrying the
// tool version, native MathML formulas, and an entry for every metric.
func TestWriteDocs(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	if err := reporting.WriteDocs(sink, "v1.2.3"); err != nil {
		t.Fatal(err)
	}

	html := sink.buf.String()
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

	// Self-containment: nothing on the page may fetch an external resource
	// — the MathML must render without any math library.
	for _, ref := range []string{`src="http`, `href="http`, `url(http`, `@import`} {
		if strings.Contains(html, ref) {
			t.Errorf("guide contains external reference %q; it must be self-contained", ref)
		}
	}
}
