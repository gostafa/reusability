// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package reusability_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cenkalti/backoff/v7"
	"github.com/gostafa/reusability"
	"github.com/gostafa/singleton"
)

func TestAnalyzeReadsChangedSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "go.mod"),
		[]byte("module example.com/fresh\n\ngo 1.24\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	config := reusability.Config{Directory: dir}
	for _, name := range []string{"Before", "After"} {
		if err := os.WriteFile(
			filepath.Join(dir, "types.go"),
			[]byte("package fresh\n\ntype "+name+" struct{}\n"),
			0o600,
		); err != nil {
			t.Fatal(err)
		}
		report, err := reusability.Analyze(t.Context(), &config)
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Packages) != 1 || len(report.Packages[0].Types) != 1 ||
			report.Packages[0].Types[0].Name != name {
			t.Fatalf("changed source was cached: %+v", report)
		}
	}
}

func TestAnalyzeErrorBoundary(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, test := range []struct {
		ctx    context.Context
		config reusability.Config
		want   error
	}{
		{ctx: ctx, config: reusability.Config{}, want: reusability.ErrAnalysis},
		{ctx: t.Context(), config: reusability.Config{Directory: filepath.Join(t.TempDir(), "missing")}, want: reusability.ErrAnalysis},
		{ctx: t.Context(), config: reusability.Config{DependencyScope: "invalid"}, want: reusability.ErrConfig},
	} {
		_, err := reusability.Analyze(test.ctx, &test.config)
		var retryErr *backoff.RetryError
		var initErr *singleton.InitError
		var pathErr *os.PathError
		if !errors.Is(err, test.want) || errors.As(err, &retryErr) || errors.As(err, &initErr) ||
			errors.As(err, &pathErr) {
			t.Fatalf("public error boundary = %v", err)
		}
	}
}
