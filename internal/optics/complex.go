package optics

import (
	"math"
	"math/cmplx"
)

const TwoPi = 2 * math.Pi

func Abs2(z complex128) float64 {
	return real(z)*real(z) + imag(z)*imag(z)
}

func IsRealish(z complex128, tol float64) bool {
	if tol <= 0 {
		tol = 1e-12
	}
	return math.Abs(imag(z)) <= tol*math.Max(1, math.Abs(real(z)))
}

func ImagPart(z complex128) float64 {
	return imag(z)
}

func HypotSqrt(x complex128) complex128 {
	return cmplx.Sqrt(x)
}

func ComplexFromIndex(index, extinction float64) complex128 {
	return complex(index, -extinction)
}
