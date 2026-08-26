package field

import (
	"testing"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/stackgen"
)

func TestLosslessPoyntingClosed(t *testing.T) {
	s, err := stackgen.SingleAR(1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	inc := model.Incidence{WavelengthNm: 550, AngleDeg: 0}
	p, err := Walk(s, inc, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	if err := LosslessFluxClosed(p, 1e-8); err != nil {
		t.Fatal(err)
	}
	res, err := optics.Solve(s, inc, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	if err := MatchesSolve(p, res, 1e-12); err != nil {
		t.Fatal(err)
	}
	if PeakIntensity(p) <= 0 {
		t.Fatal("peak |E|²")
	}
}

func TestAbsorbingFluxDropsThroughMetal(t *testing.T) {
	s := model.Stack{
		Incident:  model.MustMaterial(1),
		Substrate: model.MustMaterial(1.5),
		Layers: model.Layers{{
			Material:    model.Material{Index: 3.0, Extinction: 3.2},
			ThicknessNm: 20,
		}},
	}
	inc := model.Incidence{WavelengthNm: 550, AngleDeg: 0}
	p, err := Walk(s, inc, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	if err := AbsorbingFluxDrops(p); err != nil {
		t.Fatal(err)
	}
	if p.Absorption < 0.1 {
		t.Fatalf("chromium-like film should absorb, A=%g", p.Absorption)
	}
}
