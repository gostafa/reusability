package outbound

import (
	"context"
	"testing"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
)

type stubSource struct{ mod string }

func (s stubSource) Load(context.Context, FactOptions) (string, []domain.PackageExtract, error) {
	return s.mod, []domain.PackageExtract{{Path: "p"}}, nil
}

var _ FactSource = stubSource{}

// White-box: the port is satisfiable and FactOptions carries the load config.
func TestFactSourceContract(t *testing.T) {
	t.Parallel()

	var src FactSource = stubSource{mod: "example.com/m"}

	mod, pkgs, err := src.Load(
		context.Background(),
		FactOptions{Patterns: []string{"./..."}, Workers: 2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if mod != "example.com/m" || len(pkgs) != 1 {
		t.Fatalf("mod=%q pkgs=%d", mod, len(pkgs))
	}
}
