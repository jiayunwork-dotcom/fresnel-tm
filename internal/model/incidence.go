package model

import (
	"fmt"
	"math"
)

type Incidence struct {
	WavelengthNm float64 `json:"wavelength_nm"`
	AngleDeg     float64 `json:"angle_deg,omitempty"`
}

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

func DefaultIncidence() Incidence {
	return Incidence{WavelengthNm: 550, AngleDeg: 0}
}

func (i Incidence) AngleRad() float64 {
	return i.AngleDeg * math.Pi / 180
}

func (i Incidence) WithWavelength(nm float64) Incidence {
	i.WavelengthNm = nm
	return i
}
