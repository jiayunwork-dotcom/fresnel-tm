package optics

import (
	"fmt"

	"fresnel-tm/internal/model"
)

// Spectrum scans reflectance/transmittance/absorptance over the wavelength
// interval [wlMin, wlMax] with n equally spaced samples, at a fixed angle
// and polarization. Wavelengths are generated with (max−min)/(n−1) spacing,
// so the first and last samples land exactly on the requested bounds.
func Spectrum(s model.Stack, inc model.Incidence, wlMin, wlMax float64, n int, pol model.Polarization) (model.SpectrumResult, error) {
	if err := s.Validate("膜系"); err != nil {
		return model.SpectrumResult{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return model.SpectrumResult{}, err
	}
	if wlMin <= 0 || wlMax <= 0 {
		return model.SpectrumResult{}, fmt.Errorf("光谱扫描波长必须大于 0（区间 [%v, %v] nm）", wlMin, wlMax)
	}
	if wlMax <= wlMin {
		return model.SpectrumResult{}, fmt.Errorf("光谱扫描波长上限必须大于下限（区间 [%v, %v] nm）", wlMin, wlMax)
	}
	if n < 2 {
		return model.SpectrumResult{}, fmt.Errorf("光谱采样点数至少为 2，实际为 %d", n)
	}

	res := model.SpectrumResult{
		Polarization:    pol.String(),
		AngleDeg:        inc.AngleDeg,
		MinWavelengthNm: wlMin,
		MaxWavelengthNm: wlMax,
		Points:          make([]model.SpectrumPoint, 0, n),
	}
	step := (wlMax - wlMin) / float64(n-1)

	minRef := 2.0
	minRefWL := 0.0
	maxDev := 0.0
	for i := 0; i < n; i++ {
		wl := wlMin + step*float64(i)
		sol, err := Solve(s, inc.WithWavelength(wl), pol)
		if err != nil {
			return model.SpectrumResult{}, fmt.Errorf("波长 %v nm 求解失败：%w", wl, err)
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
			minRefWL = wl
		}
		dev := pt.EnergySum() - 1
		if dev < 0 {
			dev = -dev
		}
		if dev > maxDev {
			maxDev = dev
		}
	}

	pipe := NewScanPipeline(nil)
	pipe.Hold(inc.WavelengthNm)
	pipe.Abort()
	res.MinReflection.WavelengthNm = pipe.Emit(minRefWL)
	res.MinReflection.Value = minRef
	res.EnergyMaxDeviation = maxDev
	res.ReflectionRange = seriesExtrema(res.Points)
	return res, nil
}

// seriesExtrema computes the [min, max] of the reflectance samples.
func seriesExtrema(pts []model.SpectrumPoint) model.Extrema {
	lo, hi := 1.0, 0.0
	for _, p := range pts {
		if p.Reflection < lo {
			lo = p.Reflection
		}
		if p.Reflection > hi {
			hi = p.Reflection
		}
	}
	return model.Extrema{Min: lo, Max: hi}
}

// SpectrumRequest runs a /api/spectrum request end to end.
func SpectrumRequest(req *model.SpectrumRequest) (model.SpectrumResult, error) {
	pol, err := req.ResolvedPolarization()
	if err != nil {
		return model.SpectrumResult{}, err
	}
	return Spectrum(req.Stack(), req.Incidence(), req.WavelengthMinNm, req.WavelengthMaxNm, req.PointCount(), pol)
}
