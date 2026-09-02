// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// White-box: the debug Stringer summarizes a package result.
func TestPackageResultString(t *testing.T) {
	t.Parallel()

	pr := PackageResult{
		Path:  "example.com/m/p",
		Types: make([]TypeResult, 1),
	}

	s := pr.String()
	for _, want := range []string{"example.com/m/p", "1 types"} {
		if !strings.Contains(s, want) {
			t.Errorf("String()=%q missing %q", s, want)
		}
	}
}

// fakeAnalyzer proves the inbound port is implementable from outside.
type fakeAnalyzer struct {
	result Result
	err    error
}

func (f fakeAnalyzer) Analyze(context.Context, *Options) (Result, error) {
	return f.result, f.err
}

// Black-box: an external Analyzer can be built and driven through the port.
func TestAnalyzerImplementable(t *testing.T) {
	t.Parallel()

	var a Analyzer = fakeAnalyzer{result: Result{ModulePath: "example.com/m"}}

	got, err := a.Analyze(context.Background(), &Options{Patterns: []string{"./..."}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ModulePath != "example.com/m" {
		t.Fatalf("ModulePath = %q", got.ModulePath)
	}

	sentinel := errors.New("boom")
	if _, err := (fakeAnalyzer{err: sentinel}).Analyze(
		context.Background(), &Options{},
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("error not propagated: %v", err)
	}
}
