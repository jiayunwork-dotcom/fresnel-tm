package band

import (
	"fmt"
	"math"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/owt"
)

type Sample struct {
	WavelengthNm float64
	R            float64
	T            float64
	A            float64
}

func Sweep(s model.Stack, angleDeg float64, pol model.Polarization, lo, hi float64, n int) ([]Sample, error) {
	if !(lo > 0) || !(hi > lo) {
		return nil, fmt.Errorf("band: wavelength window must satisfy 0 < lo < hi")
	}
	if n < 2 {
		return nil, fmt.Errorf("band: need at least 2 samples")
	}
	out := make([]Sample, n)
	for i := 0; i < n; i++ {
		wl := lo + (hi-lo)*float64(i)/float64(n-1)
		inc := model.Incidence{WavelengthNm: wl, AngleDeg: angleDeg}
		res, err := optics.Solve(s, inc, pol)
		if err != nil {
			return nil, err
		}
		out[i] = Sample{WavelengthNm: wl, R: res.Reflection, T: res.Transmission, A: res.Absorption}
	}
	return out, nil
}

func MeanR(pts []Sample) (float64, error) {
	if len(pts) < 2 {
		return 0, fmt.Errorf("band: empty sweep")
	}
	area := 0.0
	span := 0.0
	for i := 1; i < len(pts); i++ {
		dx := pts[i].WavelengthNm - pts[i-1].WavelengthNm
		if dx <= 0 {
			return 0, fmt.Errorf("band: wavelengths must increase")
		}
		area += 0.5 * (pts[i].R + pts[i-1].R) * dx
		span += dx
	}
	return area / span, nil
}

func MinR(pts []Sample) (Sample, bool) {
	var best Sample
	found := false
	for _, p := range pts {
		if !found || p.R < best.R {
			best = p
			found = true
		}
	}
	return best, found
}

func EnergyClosed(pts []Sample, tol float64) error {
	for _, p := range pts {
		sum := p.R + p.T + p.A
		if math.Abs(sum-1) > tol {
			return fmt.Errorf("band: R+T+A=%g at %g nm", sum, p.WavelengthNm)
		}
	}
	return nil
}

func WidthBelow(pts []Sample, threshold float64) (float64, error) {
	if len(pts) < 2 {
		return 0, fmt.Errorf("band: empty sweep")
	}
	w := 0.0
	for i := 1; i < len(pts); i++ {
		dx := pts[i].WavelengthNm - pts[i-1].WavelengthNm
		if pts[i].R < threshold && pts[i-1].R < threshold {
			w += dx
		} else if (pts[i].R < threshold) != (pts[i-1].R < threshold) {
			w += 0.5 * dx
		}
	}
	return w, nil
}

func DesignDip(s model.Stack, designNm, angleDeg float64, pol model.Polarization, span float64, n int) (Sample, error) {
	if span <= 0 {
		return Sample{}, fmt.Errorf("band: span must be > 0")
	}
	pts, err := Sweep(s, angleDeg, pol, designNm-span, designNm+span, n)
	if err != nil {
		return Sample{}, err
	}
	best, ok := MinR(pts)
	if !ok {
		return Sample{}, fmt.Errorf("band: no samples")
	}
	return best, nil
}

func MaxR(pts []Sample) (Sample, bool) {
	var best Sample
	found := false
	for _, p := range pts {
		if !found || p.R > best.R {
			best = p
			found = true
		}
	}
	return best, found
}

func Stopband(pts []Sample, threshold float64) (lo, hi float64, ok bool) {
	started := false
	for _, p := range pts {
		if p.R >= threshold {
			if !started {
				lo = p.WavelengthNm
				started = true
			}
			hi = p.WavelengthNm
		} else if started {
			break
		}
	}
	return lo, hi, started && hi > lo
}

func Center(lo, hi float64) float64 {
	return 0.5 * (lo + hi)
}

func BlueShift(normalCenter, obliqueCenter float64) float64 {
	return normalCenter - obliqueCenter
}

func MatchesOWTStopband(nH, nL, designNm float64, pts []Sample) error {
	wantLo, wantHi, err := owt.StopbandEdges(nH, nL, designNm)
	if err != nil {
		return err
	}
	lo, hi, ok := Stopband(pts, 0.85)
	if !ok {
		return fmt.Errorf("band: no high-R stopband")
	}
	c := Center(lo, hi)
	wantC := Center(wantLo, wantHi)
	if math.Abs(c-wantC) > 0.12*designNm {
		return fmt.Errorf("band: stopband center %g vs owt %g", c, wantC)
	}
	return nil
}
