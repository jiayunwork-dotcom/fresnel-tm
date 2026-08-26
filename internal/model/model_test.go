package model

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestValidateMaterialRejectsBadIndex(t *testing.T) {
	cases := []struct {
		name string
		m    Material
		want string
	}{
		{"zero index", Material{Index: 0}, "折射率"},
		{"negative index", Material{Index: -1.5}, "折射率"},
		{"negative extinction", Material{Index: 1.5, Extinction: -0.1}, "消光"},
		{"nan index", Material{Index: nan()}, "有限"},
		{"inf extinction", Material{Index: 1.5, Extinction: inf()}, "有限"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.m.Validate("材料")
			if err == nil {
				t.Fatalf("Validate() = nil, want error mentioning %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err.Error(), c.want)
			}
		})
	}
}

func TestValidateLayerRejectsNegativeThickness(t *testing.T) {
	l := Layer{Material: Material{Index: 1.38}, ThicknessNm: -10}
	err := l.Validate("膜系")
	if err == nil {
		t.Fatal("Validate() = nil, want error for negative thickness")
	}
	if !strings.Contains(err.Error(), "-10") {
		t.Errorf("error %q does not report the offending thickness -10", err.Error())
	}

	zero := Layer{Material: Material{Index: 1.38}, ThicknessNm: 0}
	if err := zero.Validate("膜系"); err != nil {
		t.Errorf("zero thickness should be legal, got error: %v", err)
	}
}

func TestValidateIncidenceRejectsBadValues(t *testing.T) {
	cases := []struct {
		name string
		inc  Incidence
	}{
		{"zero wavelength", Incidence{WavelengthNm: 0, AngleDeg: 0}},
		{"negative wavelength", Incidence{WavelengthNm: -400, AngleDeg: 0}},
		{"angle 90 degrees", Incidence{WavelengthNm: 550, AngleDeg: 90}},
		{"angle above 90", Incidence{WavelengthNm: 550, AngleDeg: 91.5}},
		{"negative angle", Incidence{WavelengthNm: 550, AngleDeg: -5}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.inc.Validate("入射"); err == nil {
				t.Errorf("Validate() = nil, want error for %s", c.name)
			}
		})
	}

	ok := Incidence{WavelengthNm: 550, AngleDeg: 89.99}
	if err := ok.Validate("入射"); err != nil {
		t.Errorf("valid incidence rejected: %v", err)
	}
}

func TestParsePolarization(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Polarization
	}{
		{"s", PolS},
		{"p", PolP},
		{"average", PolAverage},
		{"avg", PolAverage},
		{"", PolAverage},
		{"S", PolS},
		{"TE", PolS},
		{"TM", PolP},
	} {
		got, err := ParsePolarization(tc.in)
		if err != nil {
			t.Errorf("ParsePolarization(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParsePolarization(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}

	if _, err := ParsePolarization("x"); err == nil {
		t.Error("ParsePolarization(\"x\") = nil error, want error")
	}
}

func TestExampleARQuarterLoads(t *testing.T) {
	data, err := os.ReadFile("../../example/ar-quarter.json")
	if err != nil {
		t.Fatalf("cannot read packaged example: %v", err)
	}
	ex, err := ParseExample(data)
	if err != nil {
		t.Fatalf("ParseExample: %v", err)
	}
	if ex.Name != "ar-quarter" {
		t.Errorf("Name = %q, want ar-quarter", ex.Name)
	}
	if len(ex.Layers) != 1 {
		t.Fatalf("layer count = %d, want 1", len(ex.Layers))
	}
	l := ex.Layers[0]
	wantD := 550.0 / (4 * 1.38)
	if diff := l.ThicknessNm - wantD; diff > 0.01 || diff < -0.01 {
		t.Errorf("thickness = %v, want quarter-wave %v (±0.01 nm)", l.ThicknessNm, wantD)
	}
}

func TestStackZeroLayersIsLegal(t *testing.T) {
	s := Stack{Incident: Material{Index: 1.0}, Substrate: Material{Index: 1.5}}
	if err := s.Validate("膜系"); err != nil {
		t.Fatalf("bare stack should validate, got: %v", err)
	}
	if !s.IsLossless() {
		t.Error("bare glass stack should be lossless")
	}
}

func nan() float64 {
	return math.NaN()
}

func inf() float64 {
	return math.Inf(1)
}
