package disp

import (
	"fmt"
	"math"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

type Cauchy struct {
	A float64
	B float64
	C float64
}

func (c Cauchy) Validate() error {
	if math.IsNaN(c.A) || math.IsInf(c.A, 0) || c.A <= 0 {
		return fmt.Errorf("disp: Cauchy A must be finite and > 0")
	}
	if math.IsNaN(c.B) || math.IsInf(c.B, 0) {
		return fmt.Errorf("disp: Cauchy B must be finite")
	}
	if math.IsNaN(c.C) || math.IsInf(c.C, 0) {
		return fmt.Errorf("disp: Cauchy C must be finite")
	}
	return nil
}

func (c Cauchy) Index(lambdaNm float64) (float64, error) {
	if err := c.Validate(); err != nil {
		return 0, err
	}
	if !(lambdaNm > 0) {
		return 0, fmt.Errorf("disp: wavelength must be > 0")
	}
	um := lambdaNm / 1000
	inv2 := 1 / (um * um)
	n := c.A + c.B*inv2 + c.C*inv2*inv2
	if !(n > 0) || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("disp: Cauchy n(%g) is not a positive finite index", lambdaNm)
	}
	return n, nil
}

type Sellmeier struct {
	Terms []SellmeierTerm
}

type SellmeierTerm struct {
	B float64
	C float64
}

func (s Sellmeier) Validate() error {
	if len(s.Terms) == 0 {
		return fmt.Errorf("disp: Sellmeier needs at least one term")
	}
	for i, t := range s.Terms {
		if math.IsNaN(t.B) || math.IsInf(t.B, 0) || t.B <= 0 {
			return fmt.Errorf("disp: Sellmeier term %d B must be > 0", i+1)
		}
		if math.IsNaN(t.C) || math.IsInf(t.C, 0) || t.C <= 0 {
			return fmt.Errorf("disp: Sellmeier term %d C must be > 0", i+1)
		}
	}
	return nil
}

func (s Sellmeier) Index(lambdaNm float64) (float64, error) {
	if err := s.Validate(); err != nil {
		return 0, err
	}
	if !(lambdaNm > 0) {
		return 0, fmt.Errorf("disp: wavelength must be > 0")
	}
	l2 := (lambdaNm / 1000) * (lambdaNm / 1000)
	sum := 1.0
	for i, t := range s.Terms {
		den := l2 - t.C
		if math.Abs(den) < 1e-15 {
			return 0, fmt.Errorf("disp: Sellmeier pole at term %d", i+1)
		}
		sum += t.B * l2 / den
	}
	if sum <= 0 {
		return 0, fmt.Errorf("disp: Sellmeier n² <= 0 at %g nm", lambdaNm)
	}
	n := math.Sqrt(sum)
	if !(n > 0) || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("disp: Sellmeier n(%g) is not finite", lambdaNm)
	}
	return n, nil
}

type Law interface {
	Index(lambdaNm float64) (float64, error)
}

func MgF2Cauchy() Cauchy {
	return Cauchy{A: 1.359, B: 0.004, C: 0.0001}
}

func GlassSellmeier() Sellmeier {
	return Sellmeier{Terms: []SellmeierTerm{{B: 1.0396, C: 0.006}, {B: 0.2318, C: 0.020}, {B: 1.0105, C: 103.56}}}
}

func ApplyFilm(s model.Stack, film Law, lambdaNm float64) (model.Stack, error) {
	if err := s.Validate("膜系"); err != nil {
		return model.Stack{}, err
	}
	n, err := film.Index(lambdaNm)
	if err != nil {
		return model.Stack{}, err
	}
	out := s
	out.Layers = make(model.Layers, len(s.Layers))
	copy(out.Layers, s.Layers)
	for i := range out.Layers {
		if out.Layers[i].Absorbs() {
			continue
		}
		out.Layers[i].Index = n
	}
	return out, nil
}

func ApplyIncidentSubstrate(s model.Stack, inc Law, sub Law, lambdaNm float64) (model.Stack, error) {
	if err := s.Validate("膜系"); err != nil {
		return model.Stack{}, err
	}
	ni, err := inc.Index(lambdaNm)
	if err != nil {
		return model.Stack{}, err
	}
	ns, err := sub.Index(lambdaNm)
	if err != nil {
		return model.Stack{}, err
	}
	out := s
	out.Incident.Index = ni
	out.Substrate.Index = ns
	return out, nil
}

func DesignQuarter(film Law, designNm float64) (model.Layer, error) {
	n, err := film.Index(designNm)
	if err != nil {
		return model.Layer{}, err
	}
	return model.Layer{
		Material:    model.MustMaterial(n),
		ThicknessNm: optics.QuarterWaveThickness(designNm, n),
	}, nil
}

func DispersedSolve(s model.Stack, film Law, inc model.Incidence, pol model.Polarization) (model.StackResult, error) {
	applied, err := ApplyFilm(s, film, inc.WavelengthNm)
	if err != nil {
		return model.StackResult{}, err
	}
	return optics.Solve(applied, inc, pol)
}

func ConstantSolve(s model.Stack, film Law, designNm float64, inc model.Incidence, pol model.Polarization) (model.StackResult, error) {
	n, err := film.Index(designNm)
	if err != nil {
		return model.StackResult{}, err
	}
	frozen := s
	frozen.Layers = make(model.Layers, len(s.Layers))
	copy(frozen.Layers, s.Layers)
	for i := range frozen.Layers {
		if frozen.Layers[i].Absorbs() {
			continue
		}
		frozen.Layers[i].Index = n
	}
	return optics.Solve(frozen, inc, pol)
}

func DispersedSpectrum(s model.Stack, film Law, angleDeg float64, pol model.Polarization, lo, hi float64, n int) (model.SpectrumResult, error) {
	if !(lo > 0) || !(hi > lo) {
		return model.SpectrumResult{}, fmt.Errorf("disp: wavelength window must satisfy 0 < lo < hi")
	}
	if n < 2 {
		return model.SpectrumResult{}, fmt.Errorf("disp: need at least 2 samples")
	}
	inc0 := model.Incidence{WavelengthNm: lo, AngleDeg: angleDeg}
	if err := inc0.Validate("入射"); err != nil {
		return model.SpectrumResult{}, err
	}
	res := model.SpectrumResult{
		Polarization:    pol.String(),
		AngleDeg:        angleDeg,
		MinWavelengthNm: lo,
		MaxWavelengthNm: hi,
		Points:          make([]model.SpectrumPoint, 0, n),
	}
	step := (hi - lo) / float64(n-1)
	minRef := 2.0
	minWL := 0.0
	maxDev := 0.0
	for i := 0; i < n; i++ {
		wl := lo + step*float64(i)
		sol, err := DispersedSolve(s, film, model.Incidence{WavelengthNm: wl, AngleDeg: angleDeg}, pol)
		if err != nil {
			return model.SpectrumResult{}, err
		}
		pt := model.SpectrumPoint{
			WavelengthNm: wl,
			Reflection:   sol.Reflection,
			Transmission: sol.Transmission,
			Absorption:   sol.Absorption,
		}
		res.Points = append(res.Points, pt)
		if pt.Reflection < minRef {
			minRef = pt.Reflection
			minWL = wl
		}
		dev := pt.EnergySum() - 1
		if dev < 0 {
			dev = -dev
		}
		if dev > maxDev {
			maxDev = dev
		}
	}
	res.MinReflection.WavelengthNm = minWL
	res.MinReflection.Value = minRef
	res.EnergyMaxDeviation = maxDev
	return res, nil
}

func ShortWaveIndexRise(film Law, shortNm, longNm float64) (float64, error) {
	ns, err := film.Index(shortNm)
	if err != nil {
		return 0, err
	}
	nl, err := film.Index(longNm)
	if err != nil {
		return 0, err
	}
	if ns <= nl {
		return 0, fmt.Errorf("disp: expected n(%g)=%g > n(%g)=%g for normal dispersion", shortNm, ns, longNm, nl)
	}
	return ns - nl, nil
}
