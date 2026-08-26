package model

import (
	"fmt"
	"math"
)

type Layer struct {
	Material
	ThicknessNm float64 `json:"thickness_nm"`
}

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

func (l Layer) OpticalThickness() float64 {
	return l.Index * l.ThicknessNm
}

func (l Layer) QuarterWaveThickness(designWavelengthNm float64) float64 {
	return designWavelengthNm / (4 * l.Index)
}

func (l Layer) PhaseDesignWavelength() float64 {
	return 4 * l.Index * l.ThicknessNm
}

type Layers []Layer

func (ls Layers) Validate(where string) error {
	for i, l := range ls {
		if err := l.Validate(fmt.Sprintf("%s 第 %d 层", where, i+1)); err != nil {
			return err
		}
	}
	return nil
}

func (ls Layers) TotalThickness() float64 {
	total := 0.0
	for _, l := range ls {
		total += l.ThicknessNm
	}
	return total
}

func (ls Layers) AbsorbingLayers() int {
	n := 0
	for _, l := range ls {
		if l.Absorbs() {
			n++
		}
	}
	return n
}
