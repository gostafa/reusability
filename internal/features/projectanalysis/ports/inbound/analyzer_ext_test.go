package inbound_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/reusability/internal/features/projectanalysis/ports/inbound"
)

// fakeAnalyzer proves the inbound port is implementable from outside.
type fakeAnalyzer struct {
	result inbound.Result
	err    error
}

func (f fakeAnalyzer) Analyze(context.Context, inbound.Options) (inbound.Result, error) {
	return f.result, f.err
}

// Black-box: an external Analyzer can be built and driven through the port.
func TestAnalyzerImplementable(t *testing.T) {
	t.Parallel()

	var a inbound.Analyzer = fakeAnalyzer{result: inbound.Result{ModulePath: "example.com/m"}}

	got, err := a.Analyze(context.Background(), inbound.Options{Patterns: []string{"./..."}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ModulePath != "example.com/m" {
		t.Fatalf("ModulePath = %q", got.ModulePath)
	}

	sentinel := errors.New("boom")
	if _, err := (fakeAnalyzer{err: sentinel}).Analyze(
		context.Background(),
		inbound.Options{},
	); !errors.Is(
		err,
		sentinel,
	) {
		t.Fatalf("error not propagated: %v", err)
	}
}
