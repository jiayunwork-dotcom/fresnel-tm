package optics

import (
	"math"
	"testing"

	"fresnel-tm/internal/model"
)

// arStack returns the packaged quarter-wave anti-reflection problem:
// air / MgF2 99.64 nm / glass, designed for 550 nm.
func arStack() model.Stack {
	return model.Stack{
		Incident:  model.MustMaterial(1.0),
		Substrate: model.MustMaterial(1.5),
		Layers:    model.Layers{{Material: model.MustMaterial(1.38), ThicknessNm: 550.0 / (4 * 1.38)}},
	}
}

// TestBareInterfaceMatchesFresnel checks the zero-layer limit: the matrix
// solver must reproduce the single-interface Fresnel formulas for both
// polarizations at normal and oblique incidence.
func TestBareInterfaceMatchesFresnel(t *testing.T) {
	s := model.Stack{Incident: model.MustMaterial(1.0), Substrate: model.MustMaterial(1.5)}
	for _, pol := range []model.Polarization{model.PolS, model.PolP} {
		for _, angle := range []float64{0, 30, 45, 60, 75} {
			inc := model.Incidence{WavelengthNm: 550, AngleDeg: angle}
			got, err := BareInterface(s, inc, pol)
			if err != nil {
				t.Fatalf("BareInterface(%v, %v°): %v", pol, angle, err)
			}
			want := FresnelReflection(1.0, 1.5, angle, pol)
			if d := math.Abs(got.Reflection - want); d > 1e-12 {
				t.Errorf("BareInterface(%v, %v°) R = %.15f, Fresnel = %.15f (diff %g)", pol, angle, got.Reflection, want, d)
			}
		}
	}
}

// TestNoAbsorptionEnergyConservation checks R + T = 1 for a lossless
// multilayer across wavelengths and angles.
func TestNoAbsorptionEnergyConservation(t *testing.T) {
	s := model.Stack{
		Incident:  model.MustMaterial(1.0),
		Substrate: model.MustMaterial(1.5),
		Layers: model.Layers{
			{Material: model.MustMaterial(1.38), ThicknessNm: 99.64},
			{Material: model.MustMaterial(2.35), ThicknessNm: 58.5},
			{Material: model.MustMaterial(1.46), ThicknessNm: 120.0},
		},
	}
	for _, pol := range []model.Polarization{model.PolS, model.PolP, model.PolAverage} {
		for _, angle := range []float64{0, 30, 45} {
			for wl := 400.0; wl <= 700; wl += 50 {
				inc := model.Incidence{WavelengthNm: wl, AngleDeg: angle}
				res, err := Solve(s, inc, pol)
				if err != nil {
					t.Fatalf("Solve(%v, %v°, %vnm): %v", pol, angle, wl, err)
				}
				if d := math.Abs(res.Reflection + res.Transmission - 1); d > ConservationTolerance {
					t.Errorf("pol=%v angle=%v° λ=%v: R+T-1 = %g, want |·| ≤ %g", pol, angle, wl, d, ConservationTolerance)
				}
				if res.Absorption < -ConservationTolerance {
					t.Errorf("pol=%v angle=%v° λ=%v: absorption = %v, want ≥ 0", pol, angle, wl, res.Absorption)
				}
			}
		}
	}
}

// TestAbsorbingStackConservation checks R + A + T = 1 with A ≥ 0 for a
// stack containing a strongly absorbing chromium layer.
func TestAbsorbingStackConservation(t *testing.T) {
	s := model.Stack{
		Incident:  model.MustMaterial(1.0),
		Substrate: model.MustMaterial(1.5),
		Layers: model.Layers{
			{Material: model.Material{Index: 3.17, Extinction: 3.33}, ThicknessNm: 20.0},
		},
	}
	for _, pol := range []model.Polarization{model.PolS, model.PolP} {
		for _, angle := range []float64{0, 30, 45} {
			res, err := Solve(s, model.Incidence{WavelengthNm: 550, AngleDeg: angle}, pol)
			if err != nil {
				t.Fatalf("Solve: %v", err)
			}
			sum := res.Reflection + res.Transmission + res.Absorption
			if d := math.Abs(sum - 1); d > ConservationTolerance {
				t.Errorf("pol=%v angle=%v°: R+A+T = %.12f, want 1 (|diff| %g ≤ %g)", pol, angle, sum, d, ConservationTolerance)
			}
			if res.Absorption < -ConservationTolerance {
				t.Errorf("pol=%v angle=%v°: absorption = %v, want ≥ 0", pol, angle, res.Absorption)
			}
		}
	}
}

// TestQuarterWaveARReducesReflection checks the core anti-reflection claim:
// at the design wavelength the quarter-wave coating has clearly lower
// reflectance than the bare substrate.
func TestQuarterWaveARReducesReflection(t *testing.T) {
	res, err := Solve(arStack(), model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	if res.Reflection >= res.BareReflection {
		t.Fatalf("coated R = %v, bare R = %v, want coated < bare", res.Reflection, res.BareReflection)
	}
	if want, got := 0.04, res.BareReflection; math.Abs(got-want) > 1e-9 {
		t.Errorf("bare R = %v, want %v", got, want)
	}
	if res.Reflection > 0.02 {
		t.Errorf("coated R = %v, want well below 2%%", res.Reflection)
	}
}

// TestHalfWaveLayerLosesAntiReflection checks the crossover rule: doubling
// the optical path of the coating to λ/2 makes the design-wavelength AR
// vanish and reflectance return to the bare-substrate value.
func TestHalfWaveLayerLosesAntiReflection(t *testing.T) {
	halfWave := arStack()
	halfWave.Layers[0].ThicknessNm = 550.0 / (2 * 1.38)

	res, err := Solve(halfWave, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolAverage)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	if d := math.Abs(res.Reflection - res.BareReflection); d > ConservationTolerance {
		t.Errorf("half-wave R = %v, bare R = %v, want equal (|diff| %g)", res.Reflection, res.BareReflection, d)
	}
	if res.Reflection <= 0.02 {
		t.Errorf("half-wave R = %v, want the AR effect gone (R well above 2%%)", res.Reflection)
	}
}

// TestGrazingAngleReflectanceIncreases checks that reflectance rises as the
// incidence approaches grazing, for both polarizations.
func TestGrazingAngleReflectanceIncreases(t *testing.T) {
	s := model.Stack{Incident: model.MustMaterial(1.0), Substrate: model.MustMaterial(1.5)}
	for _, pol := range []model.Polarization{model.PolS, model.PolP} {
		n0, err := BareInterface(s, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, pol)
		if err != nil {
			t.Fatal(err)
		}
		n85, err := BareInterface(s, model.Incidence{WavelengthNm: 550, AngleDeg: 85}, pol)
		if err != nil {
			t.Fatal(err)
		}
		if n85.Reflection <= n0.Reflection {
			t.Errorf("%v: R(85°) = %v, R(0°) = %v, want grazing higher", pol, n85.Reflection, n0.Reflection)
		}
		if n85.Reflection < 0.4 {
			t.Errorf("%v: R(85°) = %v, want a substantial rise near grazing", pol, n85.Reflection)
		}
	}
}

// TestObliqueARMinimumBlueShifts is the hard surface: the phase thickness
// must contain cosθ, so the anti-reflection minimum of the quarter-wave
// coating moves toward shorter wavelengths as the angle grows. Without the
// cosθ factor the minimum would stay pinned at the normal-incidence design
// wavelength (≈550 nm) and this test fails.
func TestObliqueARMinimumBlueShifts(t *testing.T) {
	res, err := Spectrum(arStack(), model.Incidence{WavelengthNm: 550, AngleDeg: 45},
		360, 700, 341, model.PolP)
	if err != nil {
		t.Fatalf("Spectrum: %v", err)
	}
	if res.MinReflection.WavelengthNm >= 500 {
		t.Errorf("oblique AR minimum at %v nm, want clearly blue-shifted below 500 nm (cosθ missing from δ keeps it near 550)",
			res.MinReflection.WavelengthNm)
	}
}

// TestSpectrumScanEndpointsAndCount checks the scan contract: the requested
// interval is sampled at the requested resolution and the endpoints land on
// the requested wavelengths.
func TestSpectrumScanEndpointsAndCount(t *testing.T) {
	res, err := Spectrum(arStack(), model.Incidence{WavelengthNm: 400, AngleDeg: 0},
		400, 700, 61, model.PolS)
	if err != nil {
		t.Fatalf("Spectrum: %v", err)
	}
	if len(res.Points) != 61 {
		t.Errorf("point count = %d, want 61", len(res.Points))
	}
	first := res.Points[0].WavelengthNm
	last := res.Points[len(res.Points)-1].WavelengthNm
	if math.Abs(first-400) > 1e-9 {
		t.Errorf("first wavelength = %v, want 400", first)
	}
	if math.Abs(last-700) > 1e-9 {
		t.Errorf("last wavelength = %v, want 700", last)
	}
	// Normal incidence quarter-wave: the minimum should sit near the design.
	if math.Abs(res.MinReflection.WavelengthNm-550) > 30 {
		t.Errorf("normal-incidence minimum at %v nm, want near 550", res.MinReflection.WavelengthNm)
	}
}

// TestPolarizationResultsDiffer checks that s and p use different admittance
// rules: at oblique incidence their reflectance differs measurably.
func TestPolarizationResultsDiffer(t *testing.T) {
	inc := model.Incidence{WavelengthNm: 550, AngleDeg: 45}
	rs, err := Solve(arStack(), inc, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	rp, err := Solve(arStack(), inc, model.PolP)
	if err != nil {
		t.Fatal(err)
	}
	if d := math.Abs(rs.Reflection - rp.Reflection); d < 1e-4 {
		t.Errorf("R_s = %v, R_p = %v at 45°, want them to differ (η rules must not be mixed)", rs.Reflection, rp.Reflection)
	}
}

// TestQuarterWaveDesignHelper checks the design recipe d = λ0/(4n) and the
// closed-form residual R = ((n0·ns − n²)/(n0·ns + n²))² against the solver.
func TestQuarterWaveDesignHelper(t *testing.T) {
	d := QuarterWaveThickness(550, 1.38)
	if want := 550.0 / (4 * 1.38); math.Abs(d-want) > 1e-12 {
		t.Errorf("QuarterWaveThickness = %v, want %v", d, want)
	}
	residual := MinimalQuarterWaveResidual(1.0, 1.5, 1.38)
	s := model.Stack{
		Incident:  model.MustMaterial(1.0),
		Substrate: model.MustMaterial(1.5),
		Layers:    model.Layers{{Material: model.MustMaterial(1.38), ThicknessNm: d}},
	}
	solved, err := Solve(s, model.Incidence{WavelengthNm: 550, AngleDeg: 0}, model.PolS)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(solved.Reflection-residual) > 1e-12 {
		t.Errorf("solver R = %v, closed-form residual = %v", solved.Reflection, residual)
	}
}
