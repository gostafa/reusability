package application_test

import (
	"context"
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/reusability/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

type fakeSource struct{}

func (fakeSource) Load(
	context.Context,
	tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return "example.com/m", []tfdomain.PackageExtract{
		{Path: "example.com/m/b", InModule: true, Types: []tfdomain.TypeExtract{{Name: "B"}}},
		{Path: "example.com/m/a", InModule: true, Types: []tfdomain.TypeExtract{{Name: "A"}}},
	}, nil
}

// Black-box: the service loads through the port and assembles deterministic,
// sorted facts with dense IDs.
func TestServiceCollect(t *testing.T) {
	t.Parallel()

	svc := typefacts.NewService(fakeSource{})

	facts, err := svc.Collect(
		context.Background(),
		tfoutbound.FactOptions{Patterns: []string{"./..."}},
	)
	if err != nil {
		t.Fatal(err)
	}

	if facts.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", facts.ModulePath)
	}

	if len(facts.Packages) != 2 || facts.Packages[0].Path != "example.com/m/a" {
		t.Fatalf("packages not sorted by path: %+v", facts.Packages)
	}

	for i, p := range facts.Packages {
		if p.ID != i {
			t.Errorf("package %q ID = %d, want %d", p.Path, p.ID, i)
		}
	}
}
