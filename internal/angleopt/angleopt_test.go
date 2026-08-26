package angleopt

import (
	"math"
	"testing"
)

func TestBrewsterKillsPAndPairSumsTo90(t *testing.T) {
	th, err := Brewster(1, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	want := math.Atan(1.5) * 180 / math.Pi
	if math.Abs(th-want) > 1e-12 {
		t.Fatalf("brewster %g want %g", th, want)
	}
	s, err := AtBrewster(1, 1.5, 550)
	if err != nil {
		t.Fatal(err)
	}
	if s.Rp > 1e-10 {
		t.Fatalf("Rp at Brewster should vanish, got %g", s.Rp)
	}
	if s.Rs <= s.Rp {
		t.Fatal("Rs should remain finite")
	}
	if err := ExternalBrewsterMatchesInternal(1, 1.5); err != nil {
		t.Fatal(err)
	}
}

func TestCriticalGlassToAir(t *testing.T) {
	th, ok, err := Critical(1.5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected TIR possible")
	}
	want := math.Asin(1/1.5) * 180 / math.Pi
	if math.Abs(th-want) > 1e-12 {
		t.Fatalf("critical %g", th)
	}
	beyond, err := BeyondCritical(50, 1.5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !beyond {
		t.Fatal("50 deg in glass should TIR into air")
	}
	airGlass, err := BeyondCritical(50, 1, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if airGlass {
		t.Fatal("air->glass has no TIR")
	}
}

func TestScanMinRpNearBrewster(t *testing.T) {
	pts, err := Scan(1, 1.5, 550, 90)
	if err != nil {
		t.Fatal(err)
	}
	best, ok := MinRp(pts)
	if !ok {
		t.Fatal("empty")
	}
	th, err := Brewster(1, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(best.AngleDeg-th) > 1.5 {
		t.Fatalf("min Rp at %g, brewster %g", best.AngleDeg, th)
	}
}

func TestTIRReflectanceUnity(t *testing.T) {
	if err := TIRUnity(1.5, 1, 50, 550); err != nil {
		t.Fatal(err)
	}
}

func TestCoatedBrewsterStillKillsP(t *testing.T) {
	if err := CoatedBrewsterStillKillsP(1.38, 550); err != nil {
		t.Fatal(err)
	}
}
