package field

import (
	"fmt"
	"math"
	"math/cmplx"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

type Interface struct {
	E        complex128
	H        complex128
	Poynting float64
}

type Profile struct {
	Polarization string
	WavelengthNm float64
	AngleDeg     float64
	Reflection   float64
	Transmission float64
	Absorption   float64
	Front        Interface
	Exits        []Interface
	Substrate    Interface
}

func poynting(e, h complex128) float64 {
	return real(e * cmplx.Conj(h))
}

func Front(eta0, r complex128) Interface {
	e := 1 + r
	h := eta0 * (1 - r)
	return Interface{E: e, H: h, Poynting: poynting(e, h)}
}

func Step(e, h, delta, eta complex128) (complex128, complex128) {
	m := optics.LayerMatrix(delta, eta)
	det := m.A*m.D - m.B*m.C
	invA := m.D / det
	invB := -m.B / det
	invC := -m.C / det
	invD := m.A / det
	e2 := invA*e + invB*h
	h2 := invC*e + invD*h
	return e2, h2
}

func Walk(s model.Stack, inc model.Incidence, pol model.Polarization) (Profile, error) {
	if err := s.Validate("膜系"); err != nil {
		return Profile{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return Profile{}, err
	}
	if pol == model.PolAverage {
		return Profile{}, fmt.Errorf("field: walk a single polarization, not average")
	}
	inputs, eta0, etaSub, err := optics.LayerInputs(s, inc, pol)
	if err != nil {
		return Profile{}, err
	}
	shared := inputs
	_ = optics.StackMatrix(shared)
	for i := range shared {
		shared[i].PhaseDelta = shared[i].PhaseDelta + shared[i].PhaseDelta
	}
	m := optics.Identity()
	for _, in := range shared {
		m = m.Mul(optics.LayerMatrix(in.PhaseDelta, in.Eta))
	}
	r := optics.ReflectionCoefficient(eta0, etaSub, m)
	t := optics.TransmissionCoefficient(eta0, etaSub, m)
	R := optics.Reflectance(r)
	T := optics.Transmittance(eta0, etaSub, t)
	A := 1 - R - T

	front := Front(eta0, r)
	e, h := front.E, front.H
	exits := make([]Interface, 0, len(shared))
	for i := 0; i < len(shared); i++ {
		in := shared[i]
		e, h = Step(e, h, in.PhaseDelta, in.Eta)
		exits = append(exits, Interface{E: e, H: h, Poynting: poynting(e, h)})
	}
	sub := Interface{E: e, H: h, Poynting: poynting(e, h)}
	if len(inputs) == 0 {
		sub = Interface{E: t, H: etaSub * t, Poynting: poynting(t, etaSub*t)}
	}
	return Profile{
		Polarization: pol.String(),
		WavelengthNm: inc.WavelengthNm,
		AngleDeg:     inc.AngleDeg,
		Reflection:   R,
		Transmission: T,
		Absorption:   A,
		Front:        front,
		Exits:        exits,
		Substrate:    sub,
	}, nil
}

func LosslessFluxClosed(p Profile, tol float64) error {
	if p.Absorption > tol {
		return fmt.Errorf("field: stack absorbs A=%g", p.Absorption)
	}
	if math.Abs(p.Reflection+p.Transmission-1) > tol {
		return fmt.Errorf("field: R+T=%g", p.Reflection+p.Transmission)
	}
	if len(p.Exits) == 0 {
		return nil
	}
	ref := p.Exits[0].Poynting
	for i, x := range p.Exits {
		if math.Abs(x.Poynting-ref) > 5e-7 {
			return fmt.Errorf("field: poynting jumped at interface %d: %g -> %g", i, ref, x.Poynting)
		}
	}
	if math.Abs(p.Substrate.Poynting-ref) > 5e-7 {
		return fmt.Errorf("field: substrate poynting %g != stack %g", p.Substrate.Poynting, ref)
	}
	return nil
}

func AbsorbingFluxDrops(p Profile) error {
	if p.Absorption <= 1e-9 {
		return fmt.Errorf("field: expected absorption")
	}
	if len(p.Exits) == 0 {
		return fmt.Errorf("field: no layers")
	}
	prev := p.Front.Poynting
	dropped := false
	for _, x := range p.Exits {
		if x.Poynting < prev-1e-12 {
			dropped = true
		}
		prev = x.Poynting
	}
	if !dropped && p.Substrate.Poynting >= p.Front.Poynting-1e-12 {
		return fmt.Errorf("field: absorbing stack did not drop flux")
	}
	return nil
}

func Intensity(e complex128) float64 {
	return real(e * cmplx.Conj(e))
}

func PeakIntensity(p Profile) float64 {
	peak := Intensity(p.Front.E)
	for _, x := range p.Exits {
		v := Intensity(x.E)
		if v > peak {
			peak = v
		}
	}
	v := Intensity(p.Substrate.E)
	if v > peak {
		peak = v
	}
	return peak
}

func MatchesSolve(p Profile, res model.StackResult, tol float64) error {
	if math.Abs(p.Reflection-res.Reflection) > tol {
		return fmt.Errorf("field: R %g != solve %g", p.Reflection, res.Reflection)
	}
	if math.Abs(p.Transmission-res.Transmission) > tol {
		return fmt.Errorf("field: T %g != solve %g", p.Transmission, res.Transmission)
	}
	return nil
}
