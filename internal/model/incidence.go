package model

import (
	"fmt"
	"math"
)

// Incidence describes the incoming light: a free-space wavelength and an
// angle of incidence measured from the surface normal.
//
// The wavelength must be positive (λ ≤ 0 is unphysical). The angle of
// incidence is restricted to [0°, 90°): exactly 90° is grazing and excluded
// because the fields no longer couple into the stack. Negative angles are
// rejected; the physics is symmetric about the normal, so a caller that
// needs an azimuth just uses |angle|.
type Incidence struct {
	// WavelengthNm is the vacuum wavelength in nanometres. Must be > 0.
	WavelengthNm float64 `json:"wavelength_nm"`
	// AngleDeg is the angle of incidence in degrees, 0 ≤ angle < 90.
	AngleDeg float64 `json:"angle_deg,omitempty"`
}

// Validate checks the two scalar invariants of an incidence specification.
// Both violations surface as errors so that a bad request can never silently
// produce a number.
func (i Incidence) Validate(where string) error {
	if math.IsNaN(i.WavelengthNm) || math.IsInf(i.WavelengthNm, 0) {
		return fmt.Errorf("%s: 波长不是有限数: %v", where, i.WavelengthNm)
	}
	if i.WavelengthNm <= 0 {
		return fmt.Errorf("%s: 波长必须大于 0 nm，实际为 %v", where, i.WavelengthNm)
	}
	if math.IsNaN(i.AngleDeg) || math.IsInf(i.AngleDeg, 0) {
		return fmt.Errorf("%s: 入射角不是有限数: %v", where, i.AngleDeg)
	}
	if i.AngleDeg < 0 {
		return fmt.Errorf("%s: 入射角不能为负，实际为 %v°", where, i.AngleDeg)
	}
	if i.AngleDeg >= 90 {
		return fmt.Errorf("%s: 入射角必须小于 90°，实际为 %v°", where, i.AngleDeg)
	}
	return nil
}

// DefaultIncidence returns the canonical illumination: 550 nm at normal
// incidence, the centre of the visible band.
func DefaultIncidence() Incidence {
	return Incidence{WavelengthNm: 550, AngleDeg: 0}
}

// AngleRad converts the angle of incidence to radians for the solver.
func (i Incidence) AngleRad() float64 {
	return i.AngleDeg * math.Pi / 180
}

// WithWavelength returns a copy of the incidence with a new wavelength. It is
// used by the spectrum scanner to sweep one parameter at a time.
func (i Incidence) WithWavelength(nm float64) Incidence {
	i.WavelengthNm = nm
	return i
}
