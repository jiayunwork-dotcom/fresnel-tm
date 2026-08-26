package stackgen

import (
	"fmt"
	"math"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

func QuarterLayer(index, designNm float64) (model.Layer, error) {
	if !(index > 0) || !(designNm > 0) {
		return model.Layer{}, fmt.Errorf("stackgen: index and wavelength must be > 0")
	}
	return model.Layer{
		Material:    model.MustMaterial(index),
		ThicknessNm: optics.QuarterWaveThickness(designNm, index),
	}, nil
}

func HalfLayer(index, designNm float64) (model.Layer, error) {
	q, err := QuarterLayer(index, designNm)
	if err != nil {
		return model.Layer{}, err
	}
	q.ThicknessNm *= 2
	return q, nil
}

func AirGlass() model.Stack {
	return model.Stack{
		Incident:  model.MustMaterial(1),
		Substrate: model.MustMaterial(1.5),
	}
}

func SingleAR(filmIndex, designNm float64) (model.Stack, error) {
	layer, err := QuarterLayer(filmIndex, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	s := AirGlass()
	s.Layers = model.Layers{layer}
	return s, nil
}

func Bragg(nH, nL float64, periods int, designNm float64) (model.Stack, error) {
	if periods < 1 {
		return model.Stack{}, fmt.Errorf("stackgen: periods must be >= 1")
	}
	h, err := QuarterLayer(nH, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	l, err := QuarterLayer(nL, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	layers := make(model.Layers, 0, 2*periods+1)
	for i := 0; i < periods; i++ {
		layers = append(layers, h, l)
	}
	layers = append(layers, h)
	s := AirGlass()
	s.Layers = layers
	return s, nil
}

func SolveAt(s model.Stack, wavelengthNm, angleDeg float64, pol model.Polarization) (model.StackResult, error) {
	inc := model.Incidence{WavelengthNm: wavelengthNm, AngleDeg: angleDeg}
	return optics.Solve(s, inc, pol)
}

func ARBeatsBare(filmIndex, designNm float64) error {
	s, err := SingleAR(filmIndex, designNm)
	if err != nil {
		return err
	}
	res, err := SolveAt(s, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	if res.Reflection >= res.BareReflection {
		return fmt.Errorf("stackgen: coated R %g not below bare %g", res.Reflection, res.BareReflection)
	}
	return nil
}

func HalfWaveCancels(filmIndex, designNm float64) error {
	h, err := HalfLayer(filmIndex, designNm)
	if err != nil {
		return err
	}
	s := AirGlass()
	s.Layers = model.Layers{h}
	res, err := SolveAt(s, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	if res.Reflection+0.002 < res.BareReflection {
		return fmt.Errorf("stackgen: λ/2 should not act as AR, R=%g bare=%g", res.Reflection, res.BareReflection)
	}
	return nil
}

func BraggRisesWithPeriods(nH, nL, designNm float64) error {
	prev := -1.0
	for p := 1; p <= 4; p++ {
		s, err := Bragg(nH, nL, p, designNm)
		if err != nil {
			return err
		}
		res, err := SolveAt(s, designNm, 0, model.PolAverage)
		if err != nil {
			return err
		}
		if res.Reflection+1e-9 < prev {
			return fmt.Errorf("stackgen: more periods lowered R: %g -> %g", prev, res.Reflection)
		}
		prev = res.Reflection
	}
	return nil
}

func TwoQuarterAdmittance(n1, n2, nSub float64) float64 {
	return (n1 * n1 / (n2 * n2)) * nSub
}

func TwoQuarterResidual(n0, n1, n2, nSub float64) float64 {
	y := TwoQuarterAdmittance(n1, n2, nSub)
	num := n0 - y
	den := n0 + y
	r := num / den
	return r * r
}

func VCoat(nOuter, nInner, designNm float64) (model.Stack, error) {
	outer, err := QuarterLayer(nOuter, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	inner, err := QuarterLayer(nInner, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	s := AirGlass()
	s.Layers = model.Layers{outer, inner}
	return s, nil
}

func VCoatBeatsSingle(nOuter, nInner, singleIndex, designNm float64) error {
	v, err := VCoat(nOuter, nInner, designNm)
	if err != nil {
		return err
	}
	one, err := SingleAR(singleIndex, designNm)
	if err != nil {
		return err
	}
	vr, err := SolveAt(v, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	sr, err := SolveAt(one, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	if vr.Reflection >= sr.Reflection {
		return fmt.Errorf("stackgen: V-coat R %g did not beat single-layer %g", vr.Reflection, sr.Reflection)
	}
	return nil
}

func HalfInnerBreaksVCoat(nOuter, nInner, designNm float64) error {
	v, err := VCoat(nOuter, nInner, designNm)
	if err != nil {
		return err
	}
	ok, err := SolveAt(v, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	broken := v
	half, err := HalfLayer(nInner, designNm)
	if err != nil {
		return err
	}
	broken.Layers = model.Layers{v.Layers[0], half}
	bad, err := SolveAt(broken, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	if bad.Reflection <= ok.Reflection+0.002 {
		return fmt.Errorf("stackgen: λ/2 inner should break V-coat, R %g vs matched %g", bad.Reflection, ok.Reflection)
	}
	one, err := SingleAR(nOuter, designNm)
	if err != nil {
		return err
	}
	sr, err := SolveAt(one, designNm, 0, model.PolAverage)
	if err != nil {
		return err
	}
	if bad.Reflection+0.002 < sr.Reflection {
		return fmt.Errorf("stackgen: broken V-coat still beats single AR, R=%g single=%g", bad.Reflection, sr.Reflection)
	}
	return nil
}

func ObliqueQuarterAR(filmIndex, designNm, angleDeg float64) (model.Stack, error) {
	if angleDeg < 0 || angleDeg >= 90 {
		return model.Stack{}, fmt.Errorf("stackgen: angle must be in [0, 90)")
	}
	c := 1.0
	if angleDeg > 0 {
		s := math.Sin(angleDeg * math.Pi / 180)
		sinFilm := s / filmIndex
		if sinFilm >= 1 {
			return model.Stack{}, fmt.Errorf("stackgen: TIR in film")
		}
		c = math.Sqrt(1 - sinFilm*sinFilm)
	}
	layer, err := QuarterLayer(filmIndex, designNm)
	if err != nil {
		return model.Stack{}, err
	}
	layer.ThicknessNm = designNm / (4 * filmIndex * c)
	s := AirGlass()
	s.Layers = model.Layers{layer}
	return s, nil
}
