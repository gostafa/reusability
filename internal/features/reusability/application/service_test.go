package application

import (
	"testing"

	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// White-box: weight defaulting and validation at construction.
func TestNewServiceDefaultsAndValidation(t *testing.T) {
	t.Parallel()

	if _, err := NewService(metrics.ReusabilityWeights{}); err != nil {
		t.Fatalf("zero weights should select defaults, got %v", err)
	}

	if _, err := NewService(metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1}); err == nil {
		t.Fatal("negative weight accepted")
	}
}

// White-box: the service delegates to the domain formula.
func TestServiceComputeForType(t *testing.T) {
	t.Parallel()

	svc, err := NewService(metrics.ReusabilityWeights{})
	if err != nil {
		t.Fatal(err)
	}

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs:         []int{1, 2},
		ExportedMembers:           2,
		DocumentedExportedMembers: 2,
	}

	got := svc.ComputeForType(tf, metrics.AMC(2, 2), metrics.LCOM(2, 2, 2))
	if got.CBO != metrics.CBO(2) {
		t.Errorf("CBO = %+v, want %+v", got.CBO, metrics.CBO(2))
	}

	if !got.Reusability.Applicable {
		t.Errorf("reusability not applicable: %s", got.Reusability.Reason)
	}
}
