package owt

import (
	"math"
	"testing"
)

func TestQuarterAndHalfRoundTrip(t *testing.T) {
	d, err := ThicknessForWaves(1.38, 550, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := IsQuarter(1.38, d, 550, 1e-12)
	if err != nil || !ok {
		t.Fatalf("quarter: %v %v", ok, err)
	}
	h, err := ThicknessForWaves(1.38, 550, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	ok, err = IsHalf(1.38, h, 550, 1e-12)
	if err != nil || !ok {
		t.Fatalf("half: %v %v", ok, err)
	}
	ph, err := Phase(1.38, d, 550, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(ph-math.Pi/2) > 1e-9 {
		t.Fatalf("phase %g want π/2", ph)
	}
	det, err := Detune(1.38, d, 550, 550)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(det) > 1e-12 {
		t.Fatalf("detune at design %g", det)
	}
	halfW, err := StopbandHalfWidth(2.3, 1.38)
	if err != nil {
		t.Fatal(err)
	}
	if halfW <= 0 {
		t.Fatal("stopband width")
	}
	c, err := FilmCosine(1, 1.38, 35)
	if err != nil {
		t.Fatal(err)
	}
	if c >= 1 {
		t.Fatal("oblique film cosine should drop")
	}
	dq, err := ThicknessForQuarterOblique(1.38, 550, c)
	if err != nil {
		t.Fatal(err)
	}
	if dq <= d {
		t.Fatalf("oblique QW thickness %g should exceed normal %g", dq, d)
	}
	lo, hi, err := StopbandEdges(2.3, 1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	if !(lo < 550 && hi > 550) {
		t.Fatalf("stopband edges %g %g", lo, hi)
	}
}
