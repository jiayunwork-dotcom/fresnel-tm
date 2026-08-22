// Package optics implements the 2×2 characteristic matrix method for
// plane-wave reflection and transmission by a stack of plane-parallel
// layers.
//
// For every layer the solver builds a matrix
//
//	M_j = [[cos δ,   i sin δ / η],
//	       [i η sin δ,      cos δ]]
//
// where δ = 2π·n̂·d·cosθ/λ is the complex phase thickness, η is the layer
// admittance (η = n̂·cosθ for s-polarization, η = n̂/cosθ for p), and n̂ is
// the complex refractive index. The stack matrix is the ordered product
// M = M_1·M_2·…·M_N. The incident and substrate admittances then give the
// complex amplitude reflection and transmission coefficients
//
//	r = (η0·M11 + η0·ηs·M12 − M21 − ηs·M22) / D
//	t = 2·η0 / D,  D = η0·M11 + η0·ηs·M12 + M21 + ηs·M22
//
// and the power fractions
//
//	R = |r|²,  T = (Re ηs / Re η0)·|t|²,  A = 1 − R − T.
//
// The transmittance carries the admittance ratio (the "angle factor"); for
// lossless stacks R + T = 1 and for absorbing stacks R + A + T = 1 with
// A ≥ 0. A stack with zero layers collapses to the single-interface Fresnel
// coefficient (η0 − ηs)/(η0 + ηs).
//
// The matrix is invariant under flipping the sign of cosθ (a flipped sign
// flips both η and δ and leaves every matrix element unchanged), so the
// principal square-root branch suffices and gives physical R/T values for
// total internal reflection and for absorbing media alike.
package optics

import (
	"math"
	"math/cmplx"
)

// TwoPi is 2π, used throughout the phase-thickness formulas.
const TwoPi = 2 * math.Pi

// Abs2 returns the squared modulus |z|² without an intermediate sqrt.
func Abs2(z complex128) float64 {
	return real(z)*real(z) + imag(z)*imag(z)
}

// IsRealish reports whether z is real within a relative tolerance. It is used
// by the solver to detect lossless subsystems without comparing against an
// absolute epsilon.
func IsRealish(z complex128, tol float64) bool {
	if tol <= 0 {
		tol = 1e-12
	}
	return math.Abs(imag(z)) <= tol*math.Max(1, math.Abs(real(z)))
}

// ImagPart returns the imaginary part of a complex number as a float.
func ImagPart(z complex128) float64 {
	return imag(z)
}

// HypotSqrt is a helper that mirrors cmplx.Sqrt with an explicit non-negative
// real-part branch. The solver relies on the principal branch everywhere;
// this function documents that intent.
func HypotSqrt(x complex128) complex128 {
	return cmplx.Sqrt(x)
}

// ComplexFromIndex builds the n̂ = N − i·K index from real parts. The minus
// sign is the optics convention used across the solver and the tests.
func ComplexFromIndex(index, extinction float64) complex128 {
	return complex(index, -extinction)
}
