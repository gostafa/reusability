package application_test

import (
	"testing"

	reusability "github.com/gostafa/reusability/internal/features/reusability/application"
	typefacts "github.com/gostafa/reusability/internal/features/typefacts/domain"
	"github.com/gostafa/reusability/internal/shared/metrics"
)

// Black-box: constructing with explicit weights and evaluating a type.
func TestServiceEndToEnd(t *testing.T) {
	t.Parallel()

	svc, err := reusability.NewService(metrics.ReusabilityWeights{
		Cohesion: 0.4, Coupling: 0.3, Testability: 0.2, Documentation: 0.1,
	})
	if err != nil {
		t.Fatal(err)
	}

	tf := &typefacts.TypeFacts{
		ReferencedTypeIDs:         []int{1, 2, 3},
		ExportedMembers:           1,
		DocumentedExportedMembers: 1,
	}

	got := svc.ComputeForType(tf, metrics.AMC(4, 2), metrics.LCOM(3, 2, 2))
	if got.CBO.Value != 3 {
		t.Errorf("CBO = %v, want 3", got.CBO.Value)
	}

	if !got.Reusability.Applicable {
		t.Errorf("reusability not applicable: %s", got.Reusability.Reason)
	}
}

// Black-box: a negative weight is rejected at construction.
func TestNewServiceRejectsNegativeWeight(t *testing.T) {
	t.Parallel()

	if _, err := reusability.NewService(
		metrics.ReusabilityWeights{Cohesion: -0.5, Coupling: 1},
	); err == nil {
		t.Fatal("negative weight accepted")
	}
}
