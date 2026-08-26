package disp

import (
	"math"
	"testing"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/stackgen"
)

func TestCauchyDecreasesWithWavelength(t *testing.T) {
	c := MgF2Cauchy()
	rise, err := ShortWaveIndexRise(c, 400, 700)
	if err != nil {
		t.Fatal(err)
	}
	if rise < 0.002 {
		t.Fatalf("expected noticeable Cauchy rise, got %g", rise)
	}
	n550, err := c.Index(550)
	if err != nil {
		t.Fatal(err)
	}
	if n550 < 1.36 || n550 > 1.40 {
		t.Fatalf("MgF2 n(550)=%g", n550)
	}
}

func TestDispersedSpectrumDipShiftsFromConstantN(t *testing.T) {
	film := MgF2Cauchy()
	layer, err := DesignQuarter(film, 550)
	if err != nil {
		t.Fatal(err)
	}
	s := stackgen.AirGlass()
	s.Layers = model.Layers{layer}

	dispSpec, err := DispersedSpectrum(s, film, 0, model.PolAverage, 400, 700, 61)
	if err != nil {
		t.Fatal(err)
	}
	constSpec, err := optics.Spectrum(s, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, 400, 700, 61, model.PolAverage)
	if err != nil {
		t.Fatal(err)
	}
	if len(dispSpec.Points) != len(constSpec.Points) {
		t.Fatalf("point count %d vs %d", len(dispSpec.Points), len(constSpec.Points))
	}
	d400 := dispSpec.Points[0].Reflection
	c400 := constSpec.Points[0].Reflection
	if math.Abs(d400-c400) < 5e-4 {
		t.Fatalf("short-wave R should move when n(λ) is applied: dispersed %g constant %g", d400, c400)
	}
	d550, err := DispersedSolve(s, film, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatal(err)
	}
	c550, err := ConstantSolve(s, film, 550, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(d550.Reflection-c550.Reflection) > 1e-12 {
		t.Fatalf("at design λ, dispersed and frozen n must match, %g vs %g", d550.Reflection, c550.Reflection)
	}
	if d550.Reflection >= d550.BareReflection {
		t.Fatalf("dispersed AR should still beat bare at design")
	}
}
