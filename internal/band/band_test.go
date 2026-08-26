package band

import (
	"math"
	"testing"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/stackgen"
)

func TestARDipNearDesignAndEnergyClosed(t *testing.T) {
	d := optics.QuarterWaveThickness(550, 1.38)
	s := model.Stack{
		Incident:  model.MustMaterial(1),
		Substrate: model.MustMaterial(1.5),
		Layers: model.Layers{{
			Material:    model.MustMaterial(1.38),
			ThicknessNm: d,
		}},
	}
	pts, err := Sweep(s, 0, model.PolAverage, 400, 700, 31)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnergyClosed(pts, 1e-9); err != nil {
		t.Fatal(err)
	}
	best, ok := MinR(pts)
	if !ok {
		t.Fatal("empty")
	}
	if math.Abs(best.WavelengthNm-550) > 40 {
		t.Fatalf("dip at %g nm, expected near 550", best.WavelengthNm)
	}
	mean, err := MeanR(pts)
	if err != nil {
		t.Fatal(err)
	}
	if mean <= best.R {
		t.Fatalf("mean R should exceed the dip")
	}
}

func TestBraggStopbandMatchesOWT(t *testing.T) {
	s, err := stackgen.Bragg(2.3, 1.38, 6, 550)
	if err != nil {
		t.Fatal(err)
	}
	pts, err := Sweep(s, 0, model.PolS, 400, 700, 61)
	if err != nil {
		t.Fatal(err)
	}
	if err := MatchesOWTStopband(2.3, 1.38, 550, pts); err != nil {
		t.Fatal(err)
	}
}

func TestObliqueStopbandBlueShifts(t *testing.T) {
	s, err := stackgen.Bragg(2.3, 1.38, 6, 550)
	if err != nil {
		t.Fatal(err)
	}
	n0, err := Sweep(s, 0, model.PolS, 420, 680, 53)
	if err != nil {
		t.Fatal(err)
	}
	n35, err := Sweep(s, 35, model.PolS, 420, 680, 53)
	if err != nil {
		t.Fatal(err)
	}
	c0, ok0 := MaxR(n0)
	c35, ok35 := MaxR(n35)
	if !ok0 || !ok35 {
		t.Fatal("empty")
	}
	shift := BlueShift(c0.WavelengthNm, c35.WavelengthNm)
	if shift < 5 {
		t.Fatalf("stopband peak should blue-shift at 35°, Δ=%g (normal %g oblique %g)", shift, c0.WavelengthNm, c35.WavelengthNm)
	}
}
