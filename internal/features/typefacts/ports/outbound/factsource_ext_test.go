package outbound_test

import (
	"context"
	"testing"

	"github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/features/typefacts/ports/outbound"
)

// fakeSource is an external adapter implementing the outbound port.
type fakeSource struct{}

func (fakeSource) Load(
	_ context.Context,
	opts outbound.FactOptions,
) (string, []domain.PackageExtract, error) {
	return "example.com/m", []domain.PackageExtract{
		{Path: "example.com/m/a", InModule: true, Types: []domain.TypeExtract{{Name: "A"}}},
	}, nil
}

// Black-box: an external FactSource can be plugged in through the port.
func TestFactSourceImplementable(t *testing.T) {
	t.Parallel()

	var src outbound.FactSource = fakeSource{}

	mod, pkgs, err := src.Load(context.Background(), outbound.FactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if mod != "example.com/m" || len(pkgs) != 1 || pkgs[0].Types[0].Name != "A" {
		t.Fatalf("unexpected extract: mod=%q pkgs=%+v", mod, pkgs)
	}
}
