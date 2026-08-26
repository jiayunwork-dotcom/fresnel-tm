package model

import (
	"errors"
	"fmt"
	"math"
)

type Material struct {
	Index      float64 `json:"index"`
	Extinction float64 `json:"extinction,omitempty"`
}

func (m Material) ComplexIndex() complex128 {
	return complex(m.Index, -m.Extinction)
}

func (m Material) Absorbs() bool {
	return m.Extinction > 0
}

func (m Material) IsLossless() bool {
	return m.Extinction == 0
}

func (m Material) Validate(where string) error {
	if math.IsNaN(m.Index) || math.IsInf(m.Index, 0) {
		return fmt.Errorf("%s: 折射率不是有限数: %v", where, m.Index)
	}
	if m.Index <= 0 {
		return fmt.Errorf("%s: 折射率实部必须 > 0，实际为 %v", where, m.Index)
	}
	if math.IsNaN(m.Extinction) || math.IsInf(m.Extinction, 0) {
		return fmt.Errorf("%s: 消光系数不是有限数: %v", where, m.Extinction)
	}
	if m.Extinction < 0 {
		return fmt.Errorf("%s: 消光系数不能为负，实际为 %v", where, m.Extinction)
	}
	return nil
}

func DefaultIncident() Material {
	return Material{Index: 1.0}
}

func MustMaterial(index float64) Material {
	m := Material{Index: index}
	if err := m.Validate("MustMaterial"); err != nil {
		panic(err)
	}
	return m
}

var ErrNoLayers = errors.New("膜系没有给出任何层")
