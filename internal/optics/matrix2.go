package optics

import (
	"math/cmplx"
)

// Matrix2 is a 2×2 complex matrix
//
//	[[A, B],
//	 [C, D]]
//
// representing the characteristic matrix of a layer or the product of
// several. The layout matches the textbook convention M = Π M_j with the
// incident side on the left.
type Matrix2 struct {
	A, B, C, D complex128
}

// Identity returns the identity matrix, which is also the characteristic
// matrix of a zero-thickness stack (no layers).
func Identity() Matrix2 {
	return Matrix2{A: 1, B: 0, C: 0, D: 1}
}

// Mul returns the matrix product m·n. Matrix multiplication is associative,
// so the stack product is accumulated left-to-right in layer order.
func (m Matrix2) Mul(n Matrix2) Matrix2 {
	return Matrix2{
		A: m.A*n.A + m.B*n.C,
		B: m.A*n.B + m.B*n.D,
		C: m.C*n.A + m.D*n.C,
		D: m.C*n.B + m.D*n.D,
	}
}

// LayerMatrix builds the characteristic matrix of a single layer:
//
//	[[cos δ,   i·sin δ / η],
//	 [i·η·sin δ,      cos δ]]
//
// with phase thickness δ and admittance η. The off-diagonal entries carry
// the 1/η and η factors that make the product over layers energy-conserving.
func LayerMatrix(phaseDelta, eta complex128) Matrix2 {
	cosD := cmplx.Cos(phaseDelta)
	sinD := cmplx.Sin(phaseDelta)
	iSinD := complex(-imag(sinD), real(sinD)) // i · sin δ
	return Matrix2{
		A: cosD,
		B: iSinD / eta,
		C: eta * iSinD,
		D: cosD,
	}
}

// PhaseThickness computes the complex phase thickness of one layer:
//
//	δ = 2π · n̂ · d · cosθ / λ
//
// The cosθ factor is decisive: forgetting it leaves oblique-incidence
// spectra pinned to the normal-incidence design wavelength instead of
// blue-shifting as the angle grows.
func PhaseThickness(n complex128, cosTheta complex128, thicknessNm, wavelengthNm float64) complex128 {
	return complex(TwoPi*thicknessNm/wavelengthNm, 0) * n * cosTheta
}

// StackMatrix multiplies the characteristic matrices of all layers in order
// and returns the total matrix. An empty layer list yields the identity.
func StackMatrix(layers []LayerInput) Matrix2 {
	total := Identity()
	for _, l := range layers {
		total = total.Mul(LayerMatrix(l.PhaseDelta, l.Eta))
	}
	return total
}

// LayerInput bundles the per-layer values needed to build one characteristic
// matrix. Keeping the tuple explicit lets the solver compute phases and
// admittances once and the tests reuse the same inputs.
type LayerInput struct {
	PhaseDelta complex128
	Eta        complex128
}
