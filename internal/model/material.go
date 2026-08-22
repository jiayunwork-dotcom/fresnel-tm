// Package model defines the data structures, validation rules and JSON wire
// format shared by the CLI, the HTTP API and the web frontend of fresnel-tm.
//
// A thin-film problem is described by three optical media:
//
//   - an incident medium (superstrate), the half-space the light comes from;
//   - an ordered list of plane-parallel layers deposited on the substrate;
//   - a substrate, the half-space the light exits into.
//
// Every medium carries a complex refractive index. The real part is the
// ordinary refractive index N; the extinction coefficient K ≥ 0 encodes
// absorption (K = 0 for lossless media such as air, MgF2 or glass). The
// solver in package optics turns a Stack plus an Incidence into reflectance,
// transmittance and absorptance values using the 2×2 characteristic matrix
// method.
//
// All lengths are expressed in nanometres and all angles in degrees. Wavelength
// and thickness use nanometres so visible light (≈380–780 nm) and quarter-wave
// coatings (≈100 nm) read as friendly numbers.
package model

import (
	"errors"
	"fmt"
	"math"
)

// Material describes one optical medium at the working wavelength.
//
// The complex refractive index is n̂ = Index − i·Extinction with the standard
// optics convention (an absorbing medium carries negative imaginary part).
type Material struct {
	// Index is the real part of the refractive index and must be > 0.
	Index float64 `json:"index"`
	// Extinction is the extinction coefficient K ≥ 0. Zero means lossless.
	Extinction float64 `json:"extinction,omitempty"`
}

// ComplexIndex returns the complex refractive index n̂ = N − i·K.
func (m Material) ComplexIndex() complex128 {
	return complex(m.Index, -m.Extinction)
}

// Absorbs reports whether the material dissipates energy at this wavelength.
func (m Material) Absorbs() bool {
	return m.Extinction > 0
}

// IsLossless reports whether the material has zero extinction coefficient.
func (m Material) IsLossless() bool {
	return m.Extinction == 0
}

// Validate checks the physical invariants of a medium. An index that is not
// finite, not positive, or an extinction coefficient below zero are rejected.
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

// DefaultIncident returns the canonical incident medium: vacuum/air, n = 1.0.
func DefaultIncident() Material {
	return Material{Index: 1.0}
}

// MustMaterial builds a lossless material, panicking on invalid input. It is
// intended for tests and example files that ship with the repository.
func MustMaterial(index float64) Material {
	m := Material{Index: index}
	if err := m.Validate("MustMaterial"); err != nil {
		panic(err)
	}
	return m
}

// ErrNoLayers is returned by callers that require at least one layer when
// none is present. The bare-interface limit (zero layers) is legal for the
// physics itself; only specific callers opt out of it.
var ErrNoLayers = errors.New("膜系没有给出任何层")
