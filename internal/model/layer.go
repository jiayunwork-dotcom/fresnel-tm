package model

import (
	"fmt"
	"math"
)

// Layer is one plane-parallel coating layer: a material plus its geometric
// thickness. The thickness is measured in nanometres along the surface
// normal. A zero thickness is allowed (a zero-thickness layer is optically
// invisible); a negative thickness is unphysical and rejected.
type Layer struct {
	Material
	// ThicknessNm is the geometric thickness in nanometres. Must be ≥ 0.
	ThicknessNm float64 `json:"thickness_nm"`
}

// Validate checks the layer invariants: a valid material and a finite,
// non-negative thickness.
func (l Layer) Validate(where string) error {
	if err := l.Material.Validate(where); err != nil {
		return err
	}
	if math.IsNaN(l.ThicknessNm) || math.IsInf(l.ThicknessNm, 0) {
		return fmt.Errorf("%s: 层厚不是有限数: %v", where, l.ThicknessNm)
	}
	if l.ThicknessNm < 0 {
		return fmt.Errorf("%s: 层厚不能为负，实际为 %v nm", where, l.ThicknessNm)
	}
	return nil
}

// OpticalThickness returns the optical path length n·d in nanometres for a
// lossless layer, or the real part n·d when the layer absorbs.
func (l Layer) OpticalThickness() float64 {
	return l.Index * l.ThicknessNm
}

// QuarterWaveThickness returns the geometric thickness in nanometres that
// makes this layer a quarter-wave plate at the given design wavelength:
//
//	d = λ0 / (4·n)
//
// At normal incidence this puts the phase thickness δ = 2π·n·d/λ0 at π/2,
// the classic single-layer anti-reflection condition.
func (l Layer) QuarterWaveThickness(designWavelengthNm float64) float64 {
	return designWavelengthNm / (4 * l.Index)
}

// PhaseDesignWavelength returns the wavelength at which a layer of the given
// thickness is exactly one quarter wave (δ = π/2) at normal incidence:
//
//	λ0 = 4·n·d
//
// This is the inverse of QuarterWaveThickness and is used by the frontend to
// label the anti-reflection dip of an example coating.
func (l Layer) PhaseDesignWavelength() float64 {
	return 4 * l.Index * l.ThicknessNm
}

// Layers is a convenience slice type for the ordered list of coatings.
type Layers []Layer

// Validate validates every layer in order and reports the index of the first
// offending layer in the error message.
func (ls Layers) Validate(where string) error {
	for i, l := range ls {
		if err := l.Validate(fmt.Sprintf("%s 第 %d 层", where, i+1)); err != nil {
			return err
		}
	}
	return nil
}

// TotalThickness sums the geometric thickness of all layers in nanometres.
func (ls Layers) TotalThickness() float64 {
	total := 0.0
	for _, l := range ls {
		total += l.ThicknessNm
	}
	return total
}

// AbsorbingLayers counts how many layers carry a non-zero extinction
// coefficient. A stack with no absorbing layers is a lossless system and
// must satisfy R + T = 1 at every wavelength.
func (ls Layers) AbsorbingLayers() int {
	n := 0
	for _, l := range ls {
		if l.Absorbs() {
			n++
		}
	}
	return n
}
