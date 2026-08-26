package optics

import (
	"math"
	"math/cmplx"
)

func SnellSine(nInc, theta0 float64, n complex128) complex128 {
	return complex(nInc*math.Sin(theta0), 0) / n
}

func RefractionCosine(nInc, theta0 float64, n complex128) complex128 {
	sin2 := SnellSine(nInc, theta0, n)
	sin2 = sin2 * sin2
	return cmplx.Sqrt(1 - sin2)
}

func IncidentCosine(angleRad float64) complex128 {
	return complex(math.Cos(angleRad), 0)
}

func LayerAngleDeg(nInc, theta0 float64, n complex128) (float64, bool) {
	s := SnellSine(nInc, theta0, n)
	c := cmplx.Sqrt(1 - s*s)
	if !IsRealish(c, 1e-9) {
		return 0, false
	}
	theta := math.Acos(clampReal(c))
	return theta * 180 / math.Pi, true
}

func clampReal(z complex128) float64 {
	re := real(z)
	if re > 1 {
		return 1
	}
	if re < -1 {
		return -1
	}
	return re
}
