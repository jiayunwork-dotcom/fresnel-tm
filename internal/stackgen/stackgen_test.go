package stackgen

import (
	"testing"

	"fresnel-tm/internal/model"
)

func TestQuarterARBeatsBareAndHalfDoesNot(t *testing.T) {
	if err := ARBeatsBare(1.38, 550); err != nil {
		t.Fatal(err)
	}
	if err := HalfWaveCancels(1.38, 550); err != nil {
		t.Fatal(err)
	}
}

func TestVCoatBeatsSingleAR(t *testing.T) {
	if err := VCoatBeatsSingle(1.38, 1.70, 1.38, 550); err != nil {
		t.Fatal(err)
	}
	res := TwoQuarterResidual(1, 1.38, 1.70, 1.5)
	if res < 0 || res > 0.02 {
		t.Fatalf("two-quarter residual %g", res)
	}
}

func TestVCoatHalfWaveLosesMatch(t *testing.T) {
	if err := HalfInnerBreaksVCoat(1.38, 1.70, 550); err != nil {
		t.Fatal(err)
	}
}

func TestBraggStopbandGrows(t *testing.T) {
	if err := BraggRisesWithPeriods(2.3, 1.38, 550); err != nil {
		t.Fatal(err)
	}
	s, err := Bragg(2.3, 1.38, 3, 550)
	if err != nil {
		t.Fatal(err)
	}
	if s.LayerCount() != 7 {
		t.Fatalf("layers %d", s.LayerCount())
	}
	res, err := SolveAt(s, 550, 0, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	if res.Reflection < 0.7 {
		t.Fatalf("3-period Bragg should be highly reflecting, R=%g", res.Reflection)
	}
}
