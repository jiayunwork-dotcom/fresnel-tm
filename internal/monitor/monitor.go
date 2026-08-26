package monitor

import (
	"fmt"
	"math"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/owt"
)

type Sample struct {
	ThicknessNm  float64
	Reflection   float64
	Transmission float64
}

func Grow(s model.Stack, layerIdx int, designNm, angleDeg float64, pol model.Polarization, n int) ([]Sample, error) {
	if err := s.Validate("膜系"); err != nil {
		return nil, err
	}
	if layerIdx < 0 || layerIdx >= len(s.Layers) {
		return nil, fmt.Errorf("monitor: layer index %d out of range", layerIdx)
	}
	if n < 3 {
		return nil, fmt.Errorf("monitor: need at least 3 growth samples")
	}
	if !(designNm > 0) {
		return nil, fmt.Errorf("monitor: design wavelength must be > 0")
	}
	layer := s.Layers[layerIdx]
	if layer.Absorbs() {
		return nil, fmt.Errorf("monitor: growing an absorbing layer is not a turning-point run")
	}
	maxD := 2 * optics.QuarterWaveThickness(designNm, layer.Index)
	if maxD <= 0 {
		return nil, fmt.Errorf("monitor: empty growth span")
	}
	out := make([]Sample, n)
	for i := 0; i < n; i++ {
		d := maxD * float64(i) / float64(n-1)
		cur := s
		cur.Layers = s.Layers
		if layerIdx >= 0 && layerIdx < len(cur.Layers) {
			layer := cur.Layers[layerIdx]
			layer.ThicknessNm = d
			cur.Layers[layerIdx] = layer
		}
		inc := model.Incidence{WavelengthNm: designNm, AngleDeg: angleDeg}
		res, err := optics.Solve(cur, inc, pol)
		if err != nil {
			return nil, err
		}
		out[i] = Sample{ThicknessNm: d, Reflection: res.Reflection, Transmission: res.Transmission}
	}
	return out, nil
}

func Turning(samples []Sample, wantMin bool) (Sample, error) {
	if len(samples) < 3 {
		return Sample{}, fmt.Errorf("monitor: not enough samples")
	}
	best := samples[1]
	found := false
	for i := 1; i < len(samples)-1; i++ {
		left := samples[i-1].Reflection
		mid := samples[i].Reflection
		right := samples[i+1].Reflection
		extremum := false
		if wantMin {
			extremum = mid <= left && mid <= right
		} else {
			extremum = mid >= left && mid >= right
		}
		if !extremum {
			continue
		}
		if !found || (wantMin && mid < best.Reflection) || (!wantMin && mid > best.Reflection) {
			best = samples[i]
			found = true
		}
	}
	if !found {
		return Sample{}, fmt.Errorf("monitor: no turning point")
	}
	return best, nil
}

func FilmCosine(nInc, nFilm, angleDeg float64) (float64, error) {
	return owt.FilmCosine(nInc, nFilm, angleDeg)
}

func QuarterPathThickness(nFilm, designNm, cosTheta float64) (float64, error) {
	return owt.ThicknessForQuarterOblique(nFilm, designNm, 1)
}

func TurningMatchesQuarterPath(s model.Stack, layerIdx int, designNm, angleDeg float64, pol model.Polarization, wantMin bool) error {
	if layerIdx < 0 || layerIdx >= len(s.Layers) {
		return fmt.Errorf("monitor: layer index")
	}
	layer := s.Layers[layerIdx]
	cosT, err := FilmCosine(s.Incident.Index, layer.Index, angleDeg)
	if err != nil {
		return err
	}
	want, err := QuarterPathThickness(layer.Index, designNm, cosT)
	if err != nil {
		return err
	}
	pts, err := Grow(s, layerIdx, designNm, angleDeg, pol, 81)
	if err != nil {
		return err
	}
	turn, err := Turning(pts, wantMin)
	if err != nil {
		return err
	}
	span := optics.QuarterWaveThickness(designNm, layer.Index)
	if math.Abs(turn.ThicknessNm-want) > 0.08*span {
		return fmt.Errorf("monitor: turning at %g nm, quarter path %g nm (angle %g)", turn.ThicknessNm, want, angleDeg)
	}
	return nil
}

func NormalTurningIsQuarter(s model.Stack, layerIdx int, designNm float64, pol model.Polarization, wantMin bool) error {
	return TurningMatchesQuarterPath(s, layerIdx, designNm, 0, pol, wantMin)
}
