package metrics

import (
	"math"
	"strings"
	"testing"
)

const epsilon = 1e-12

func assertApplicable(t *testing.T, r MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("%s: not applicable (%s), want value %v", r.Name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s: value %v, want %v", r.Name, r.Value, want)
	}
}

func assertNotApplicable(t *testing.T, r MetricResult) {
	t.Helper()

	if r.Applicable {
		t.Fatalf("%s: applicable with value %v, want not applicable", r.Name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s: not applicable without a reason", r.Name)
	}

	if r.Definition == "" {
		t.Fatalf("%s: missing definition", r.Name)
	}
}

func TestAMC(t *testing.T) {
	assertNotApplicable(t, AMC(0, 0))
	assertApplicable(t, AMC(5, 2), 2.5)
	assertApplicable(t, AMC(1, 1), 1)
}

func TestLCOM(t *testing.T) {
	assertNotApplicable(t, LCOM(0, 0, 3))
	assertNotApplicable(t, LCOM(0, 3, 0))
	assertApplicable(t, LCOM(3, 3, 3), 1-1.0/3)
	assertApplicable(t, LCOM(9, 3, 3), 0) // full matrix
	assertApplicable(t, LCOM(1, 1, 1), 0) // defined at a single method
	assertApplicable(t, LCOM(0, 2, 4), 1) // empty matrix
}

func TestTCC(t *testing.T) {
	assertNotApplicable(t, TCC(0, 0))
	assertNotApplicable(t, TCC(0, 1))
	assertApplicable(t, TCC(1, 3), 1.0/3)
	assertApplicable(t, TCC(3, 3), 1)
	assertApplicable(t, TCC(0, 2), 0)
}

func TestCBO(t *testing.T) {
	assertApplicable(t, CBO(0), 0) // always applicable, even at zero
	assertApplicable(t, CBO(7), 7)
}

func TestReusabilityAllComponents(t *testing.T) {
	weights := DefaultReusabilityWeights()
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentCoupling, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentTestability, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentDocumentation, Value: 1, Applicable: true},
		weights,
	)
	assertApplicable(t, r, 1)

	if r.Reason != "" {
		t.Fatalf("no dropped components expected, got reason %q", r.Reason)
	}
}

func TestReusabilityRenormalization(t *testing.T) {
	weights := DefaultReusabilityWeights()
	// Cohesion dropped: remaining weights 0.25+0.25+0.15 renormalize to 1.
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion},
		ReusabilityComponent{Name: ComponentCoupling, Value: 0.5, Applicable: true},
		ReusabilityComponent{Name: ComponentTestability, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentDocumentation, Value: 0, Applicable: true},
		weights,
	)
	want := (0.25*0.5 + 0.25*1 + 0.15*0) / 0.65
	assertApplicable(t, r, want)

	if !strings.Contains(r.Reason, ComponentCohesion) {
		t.Fatalf("reason %q does not list the dropped component", r.Reason)
	}
}

func TestReusabilityAllDropped(t *testing.T) {
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion, Reason: "type has no methods"},
		ReusabilityComponent{Name: ComponentCoupling, Reason: "no dependency data"},
		ReusabilityComponent{Name: ComponentTestability, Reason: "type has no methods"},
		ReusabilityComponent{Name: ComponentDocumentation, Reason: "type has no exported members"},
		DefaultReusabilityWeights(),
	)
	assertNotApplicable(t, r)
	// The reason names every dropped component with its own cause.
	for _, want := range []string{
		"every component dropped",
		"cohesion (type has no methods)",
		"coupling (no dependency data)",
		"testability (type has no methods)",
		"documentation (type has no exported members)",
	} {
		if !strings.Contains(r.Reason, want) {
			t.Fatalf("reason %q missing %q", r.Reason, want)
		}
	}
}

func TestReusabilityZeroWeightApplicable(t *testing.T) {
	// Applicable cohesion carries zero weight; the rest are dropped → weightSum 0
	// but not every component was dropped.
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentCoupling, Reason: "no dependency data"},
		ReusabilityComponent{Name: ComponentTestability, Reason: "type has no methods"},
		ReusabilityComponent{Name: ComponentDocumentation, Reason: "type has no exported members"},
		ReusabilityWeights{Cohesion: 0, Coupling: 0.25, Testability: 0.25, Documentation: 0.15},
	)
	assertNotApplicable(t, r)
	if !strings.Contains(r.Reason, "zero total weight") {
		t.Fatalf("reason %q missing zero-weight message", r.Reason)
	}
}

func TestReusabilityComponents(t *testing.T) {
	if c := CohesionComponent(
		LCOM(3, 3, 3),
	); !c.Applicable ||
		math.Abs(c.Value-1.0/3) > epsilon {
		t.Fatalf("cohesion component = %+v", c)
	}

	if c := CohesionComponent(LCOM(0, 0, 3)); c.Applicable {
		t.Fatalf("cohesion component should drop when LCOM is n/a")
	}

	if c := CouplingComponent(0); !c.Applicable || c.Value != 1 {
		t.Fatalf("coupling component at CBO=0 = %+v", c)
	}

	if c := CouplingComponent(3); math.Abs(c.Value-0.25) > epsilon {
		t.Fatalf("coupling component at CBO=3 = %+v", c)
	}

	if c := TestabilityComponent(AMC(1, 1)); !c.Applicable || c.Value != 1 {
		t.Fatalf("testability component at AMC=1 = %+v", c)
	}

	if c := TestabilityComponent(AMC(3, 1)); math.Abs(c.Value-1.0/3) > epsilon {
		t.Fatalf("testability component at AMC=3 = %+v", c)
	}

	if c := TestabilityComponent(AMC(0, 0)); c.Applicable {
		t.Fatalf("testability component should drop when AMC is n/a")
	}

	if c := DocumentationComponent(2, 4); !c.Applicable || c.Value != 0.5 {
		t.Fatalf("documentation component = %+v", c)
	}

	if c := DocumentationComponent(0, 0); c.Applicable {
		t.Fatalf("documentation component should drop with no exported members")
	}
}

func TestWeightsValidate(t *testing.T) {
	err := DefaultReusabilityWeights().Validate()
	if err != nil {
		t.Fatal(err)
	}

	err = (ReusabilityWeights{Cohesion: -1}).Validate()
	if err == nil {
		t.Fatal("negative weight accepted")
	}

	err = (ReusabilityWeights{}).Validate()
	if err == nil {
		t.Fatal("all-zero weights accepted")
	}
}

func TestClosure(t *testing.T) {
	got := Closure([]string{MetricReusability})

	want := []string{MetricAMC, MetricLCOM, MetricCBO, MetricReusability}
	if len(got) != len(want) {
		t.Fatalf("closure = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("closure = %v, want %v", got, want)
		}
	}

	if got := Closure([]string{MetricTCC}); len(got) != 1 || got[0] != MetricTCC {
		t.Fatalf("closure(tcc) = %v", got)
	}

	got = Closure([]string{MetricReusability, MetricAMC})
	want = []string{MetricAMC, MetricLCOM, MetricCBO, MetricReusability}
	if len(got) != len(want) {
		t.Fatalf("closure with duplicate deps = %v, want %v", got, want)
	}
}
