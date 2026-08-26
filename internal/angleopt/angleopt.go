package angleopt

import (
	"fmt"
	"math"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

func Brewster(nInc, nSub float64) (float64, error) {
	if !(nInc > 0) || !(nSub > 0) {
		return 0, fmt.Errorf("angleopt: indices must be positive")
	}
	return math.Atan(nSub/nInc) * 180 / math.Pi, nil
}

func Critical(nInc, nSub float64) (float64, bool, error) {
	if !(nInc > 0) || !(nSub > 0) {
		return 0, false, fmt.Errorf("angleopt: indices must be positive")
	}
	if nSub >= nInc {
		return 0, false, nil
	}
	return math.Asin(nSub/nInc) * 180 / math.Pi, true, nil
}

func BeyondCritical(angleDeg, nInc, nSub float64) (bool, error) {
	th, ok, err := Critical(nInc, nSub)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return angleDeg > th, nil
}

type PolarSplit struct {
	AngleDeg float64
	Rs       float64
	Rp       float64
	Ratio    float64
}

func Split(nInc, nSub, angleDeg, wavelengthNm float64) (PolarSplit, error) {
	if angleDeg < 0 || angleDeg >= 90 {
		return PolarSplit{}, fmt.Errorf("angleopt: angle must be in [0, 90)")
	}
	rs := optics.FresnelReflection(nInc, nSub, angleDeg, model.PolS)
	rp := optics.FresnelReflection(nInc, nSub, angleDeg, model.PolP)
	ratio := 0.0
	if rs > 0 {
		ratio = rp / rs
	}
	return PolarSplit{AngleDeg: angleDeg, Rs: rs, Rp: rp, Ratio: ratio}, nil
}

func AtBrewster(nInc, nSub, wavelengthNm float64) (PolarSplit, error) {
	th, err := Brewster(nInc, nSub)
	if err != nil {
		return PolarSplit{}, err
	}
	return Split(nInc, nSub, th, wavelengthNm)
}

func Scan(nInc, nSub, wavelengthNm float64, n int) ([]PolarSplit, error) {
	if n < 2 {
		return nil, fmt.Errorf("angleopt: need at least 2 samples")
	}
	out := make([]PolarSplit, n)
	for i := 0; i < n; i++ {
		ang := 89.0 * float64(i) / float64(n-1)
		s, err := Split(nInc, nSub, ang, wavelengthNm)
		if err != nil {
			return nil, err
		}
		out[i] = s
	}
	return out, nil
}

func MinRp(points []PolarSplit) (PolarSplit, bool) {
	var best PolarSplit
	found := false
	for _, p := range points {
		if !found || p.Rp < best.Rp {
			best = p
			found = true
		}
	}
	return best, found
}

func Contrast(s PolarSplit) float64 {
	if s.Rs+s.Rp == 0 {
		return 0
	}
	return (s.Rs - s.Rp) / (s.Rs + s.Rp)
}

func ExternalBrewsterMatchesInternal(nInc, nSub float64) error {
	a, err := Brewster(nInc, nSub)
	if err != nil {
		return err
	}
	b, err := Brewster(nSub, nInc)
	if err != nil {
		return err
	}
	sum := a + b
	if math.Abs(sum-90) > 1e-9 {
		return fmt.Errorf("angleopt: Brewster pair %g+%g != 90", a, b)
	}
	return nil
}

func TIRSplit(nInc, nSub, angleDeg, wavelengthNm float64) (PolarSplit, error) {
	beyond, err := BeyondCritical(angleDeg, nInc, nSub)
	if err != nil {
		return PolarSplit{}, err
	}
	if !beyond {
		return PolarSplit{}, fmt.Errorf("angleopt: angle %g is not beyond critical", angleDeg)
	}
	s := model.Stack{
		Incident:  model.MustMaterial(nInc),
		Substrate: model.MustMaterial(nSub),
	}
	inc := model.Incidence{WavelengthNm: wavelengthNm, AngleDeg: angleDeg}
	rs, err := optics.Solve(s, inc, model.PolS)
	if err != nil {
		return PolarSplit{}, err
	}
	rp, err := optics.Solve(s, inc, model.PolP)
	if err != nil {
		return PolarSplit{}, err
	}
	ratio := 0.0
	if rs.Reflection > 0 {
		ratio = rp.Reflection / rs.Reflection
	}
	return PolarSplit{AngleDeg: angleDeg, Rs: rs.Reflection, Rp: rp.Reflection, Ratio: ratio}, nil
}

func TIRUnity(nInc, nSub, angleDeg, wavelengthNm float64) error {
	sp, err := TIRSplit(nInc, nSub, angleDeg, wavelengthNm)
	if err != nil {
		return err
	}
	if math.Abs(sp.Rs-1) > 1e-6 || math.Abs(sp.Rp-1) > 1e-6 {
		return fmt.Errorf("angleopt: TIR R not unity Rs=%g Rp=%g", sp.Rs, sp.Rp)
	}
	return nil
}

func CoatedBrewsterStillKillsP(filmIndex, designNm float64) error {
	th, err := Brewster(1, 1.5)
	if err != nil {
		return err
	}
	bare, err := Split(1, 1.5, th, designNm)
	if err != nil {
		return err
	}
	if bare.Rp > 1e-9 {
		return fmt.Errorf("angleopt: bare Brewster Rp=%g", bare.Rp)
	}
	layer, err := func() (model.Layer, error) {
		if !(filmIndex > 0) || !(designNm > 0) {
			return model.Layer{}, fmt.Errorf("angleopt: film and design must be > 0")
		}
		return model.Layer{
			Material:    model.MustMaterial(filmIndex),
			ThicknessNm: optics.QuarterWaveThickness(designNm, filmIndex),
		}, nil
	}()
	if err != nil {
		return err
	}
	s := model.Stack{
		Incident:  model.MustMaterial(1),
		Substrate: model.MustMaterial(1.5),
		Layers:    model.Layers{layer},
	}
	inc := model.Incidence{WavelengthNm: designNm, AngleDeg: th}
	res, err := optics.Solve(s, inc, model.PolP)
	if err != nil {
		return err
	}
	if res.Reflection >= bare.Rs {
		return fmt.Errorf("angleopt: coated Brewster Rp %g >= bare Rs %g", res.Reflection, bare.Rs)
	}
	return nil
}
