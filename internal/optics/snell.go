package optics

import (
	"math"
	"math/cmplx"
)

// Snell sine describes how the sine of the ray angle transforms across an
// interface governed by the generalized (complex) Snell law:
//
//	sinθ_j = n_inc · sinθ_0 / n̂_j
//
// nInc is the real index of the incident medium, theta0 the incidence angle
// in radians, and n the complex index of the destination medium. The result
// is complex as soon as the destination medium absorbs or the ray is beyond
// the critical angle.
func SnellSine(nInc, theta0 float64, n complex128) complex128 {
	return complex(nInc*math.Sin(theta0), 0) / n
}

// RefractionCosine returns cosθ in the destination medium. It is the
// principal square root of 1 − sin²θ:
//
//	cosθ = √(1 − (n_inc·sinθ0/n̂)²)
//
// For a lossless destination below the critical angle this is a real cosine
// in (0, 1]. For total internal reflection or for absorbing media the value
// is genuinely complex; the sign choice is immaterial to R/T (the layer
// matrix is invariant under cosθ → −cosθ), so the principal branch is used.
func RefractionCosine(nInc, theta0 float64, n complex128) complex128 {
	sin2 := SnellSine(nInc, theta0, n)
	sin2 = sin2 * sin2
	return cmplx.Sqrt(1 - sin2)
}

// IncidentCosine returns cosθ of the incident ray in the real incident
// medium. The incidence angle is bounded to [0°, 90°) by validation, so the
// cosine is always a real number in (0, 1].
func IncidentCosine(angleRad float64) complex128 {
	return complex(math.Cos(angleRad), 0)
}

// LayerAngleDeg reports the ray angle (degrees) inside a layer with complex
// index n, assuming it is not beyond the critical angle. For absorbing
// media or TIR the "angle" is meaningless and the function reports false.
// It is a reporting helper for the CLI and the web frontend, not part of the
// solver path.
func LayerAngleDeg(nInc, theta0 float64, n complex128) (float64, bool) {
	s := SnellSine(nInc, theta0, n)
	c := cmplx.Sqrt(1 - s*s)
	if !IsRealish(c, 1e-9) {
		return 0, false
	}
	theta := math.Acos(clampReal(c))
	return theta * 180 / math.Pi, true
}

// clampReal bounds a near-unit cosine to [-1, 1] before acos, protecting the
// helper against rounding that pushes the argument outside the acos domain.
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
