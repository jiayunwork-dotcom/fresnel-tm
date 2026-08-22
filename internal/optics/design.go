package optics

import (
	"math"
)

// QuarterWaveThickness returns the geometric thickness d = λ0/(4·n) that
// makes a layer of index n a quarter wave at λ0. This is the classic
// single-layer anti-reflection recipe: at the design wavelength the phase
// thickness reaches δ = π/2 and the layer reflectance drops to
//
//	R = ((n0·ns − n²) / (n0·ns + n²))²
//
// which vanishes exactly when n² = n0·ns (the optimal index).
func QuarterWaveThickness(designWavelengthNm, layerIndex float64) float64 {
	return designWavelengthNm / (4 * layerIndex)
}

// OptimalSingleLayerIndex returns the index that makes a quarter-wave single
// layer perfectly antireflective: n = √(n0·ns). For air (1.0) on glass
// (1.5) the ideal index is about 1.225; MgF2 at 1.38 is the closest common
// coating and leaves a small residual reflectance.
func OptimalSingleLayerIndex(nInc, nSub float64) float64 {
	return math.Sqrt(nInc * nSub)
}

// QuarterWavePhaseDesign returns the wavelength at which a layer of thickness
// d and index n reaches δ = π/2 at normal incidence: λ0 = 4·n·d. This is the
// "design wavelength" label shown on spectra of quarter-wave stacks.
func QuarterWavePhaseDesign(layerIndex, thicknessNm float64) float64 {
	return 4 * layerIndex * thicknessNm
}

// MinimalQuarterWaveResidual returns the reflectance of a perfect quarter-wave
// single layer at its design wavelength, i.e. the residual R after the
// coating does its best:
//
//	R = ((n0·ns − n²) / (n0·ns + n²))²
//
// With the optimal index this is zero; with a real coating it is a small
// positive number the frontend can compare against the bare substrate.
func MinimalQuarterWaveResidual(nInc, nSub, layerIndex float64) float64 {
	num := nInc*nSub - layerIndex*layerIndex
	den := nInc*nSub + layerIndex*layerIndex
	return (num / den) * (num / den)
}

// BrewsterAngle returns the angle of incidence (degrees) at which p-
// polarized light is not reflected by a bare lossless interface:
// θ_B = atan(ns / n0). Only defined when n0 and ns are both real.
func BrewsterAngle(nInc, nSub float64) float64 {
	return math.Atan(nSub/nInc) * 180 / math.Pi
}
