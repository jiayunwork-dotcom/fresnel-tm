package optics

import (
	"math"

	"fresnel-tm/internal/model"
)

type BareSolution struct {
	Reflection   float64
	Transmission float64
	Absorption   float64
}

func BareInterface(s model.Stack, inc model.Incidence, pol model.Polarization) (BareSolution, error) {
	if err := s.Validate("膜系"); err != nil {
		return BareSolution{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return BareSolution{}, err
	}
	bare := s.BareStack()
	inputs, eta0, etaSub, err := LayerInputs(bare, inc, pol)
	if err != nil {
		return BareSolution{}, err
	}
	m := StackMatrix(inputs)
	r := ReflectionCoefficient(eta0, etaSub, m)
	t := TransmissionCoefficient(eta0, etaSub, m)
	R := Reflectance(r)
	T := Transmittance(eta0, etaSub, t)
	return BareSolution{Reflection: R, Transmission: T, Absorption: 1 - R - T}, nil
}

func BareAt(nInc, nSub float64, angleDeg, wavelengthNm float64, pol model.Polarization) float64 {
	s := model.Stack{
		Incident:  model.MustMaterial(nInc),
		Substrate: model.MustMaterial(nSub),
	}
	inc := model.Incidence{WavelengthNm: wavelengthNm, AngleDeg: angleDeg}
	bare, err := BareInterface(s, inc, pol)
	if err != nil {
		return 1
	}
	return bare.Reflection
}

func FresnelReflection(nInc, nSub, angleDeg float64, pol model.Polarization) float64 {
	theta0 := angleDeg * math.Pi / 180
	sinS := nInc * math.Sin(theta0) / nSub
	cosS := math.Sqrt(1 - sinS*sinS)
	var r float64
	if pol == model.PolP {
		r = (nSub*math.Cos(theta0) - nInc*cosS) / (nSub*math.Cos(theta0) + nInc*cosS)
	} else {
		r = (nInc*math.Cos(theta0) - nSub*cosS) / (nInc*math.Cos(theta0) + nSub*cosS)
	}
	return r * r
}
