package optics

import (
	"math"
)

func QuarterWaveThickness(designWavelengthNm, layerIndex float64) float64 {
	return designWavelengthNm / (4 * layerIndex)
}

func OptimalSingleLayerIndex(nInc, nSub float64) float64 {
	return math.Sqrt(nInc * nSub)
}

func QuarterWavePhaseDesign(layerIndex, thicknessNm float64) float64 {
	return 4 * layerIndex * thicknessNm
}

func MinimalQuarterWaveResidual(nInc, nSub, layerIndex float64) float64 {
	num := nInc*nSub - layerIndex*layerIndex
	den := nInc*nSub + layerIndex*layerIndex
	return (num / den) * (num / den)
}

func BrewsterAngle(nInc, nSub float64) float64 {
	return math.Atan(nSub/nInc) * 180 / math.Pi
}
