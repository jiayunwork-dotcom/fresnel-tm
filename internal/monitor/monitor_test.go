package monitor

import (
	"math"
	"testing"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/owt"
	"fresnel-tm/internal/stackgen"
)

func TestNormalTurningIsQuarter(t *testing.T) {
	s, err := stackgen.SingleAR(1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	if err := NormalTurningIsQuarter(s, 0, 550, model.PolAverage, true); err != nil {
		t.Fatal(err)
	}
}

func TestObliqueTurningMatchesQuarterPath(t *testing.T) {
	s, err := stackgen.SingleAR(1.38, 550)
	if err != nil {
		t.Fatal(err)
	}
	if err := TurningMatchesQuarterPath(s, 0, 550, 35, model.PolS, true); err != nil {
		t.Fatal(err)
	}
	cosT, err := owt.FilmCosine(1, 1.38, 35)
	if err != nil {
		t.Fatal(err)
	}
	oblique, err := owt.ThicknessForQuarterOblique(1.38, 550, cosT)
	if err != nil {
		t.Fatal(err)
	}
	normal := optics.QuarterWaveThickness(550, 1.38)
	if math.Abs(oblique-normal) < 1 {
		t.Fatalf("oblique QW path should thicken vs normal: %g vs %g", oblique, normal)
	}
}
