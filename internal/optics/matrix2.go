package optics

import (
	"math/cmplx"
)

type Matrix2 struct {
	A, B, C, D complex128
}

func Identity() Matrix2 {
	return Matrix2{A: 1, B: 0, C: 0, D: 1}
}

func (m Matrix2) Mul(n Matrix2) Matrix2 {
	return Matrix2{
		A: m.A*n.A + m.B*n.C,
		B: m.A*n.B + m.B*n.D,
		C: m.C*n.A + m.D*n.C,
		D: m.C*n.B + m.D*n.D,
	}
}

func LayerMatrix(phaseDelta, eta complex128) Matrix2 {
	cosD := cmplx.Cos(phaseDelta)
	sinD := cmplx.Sin(phaseDelta)
	iSinD := complex(-imag(sinD), real(sinD))
	return Matrix2{
		A: cosD,
		B: iSinD / eta,
		C: eta * iSinD,
		D: cosD,
	}
}

func PhaseThickness(n complex128, cosTheta complex128, thicknessNm, wavelengthNm float64) complex128 {
	return complex(TwoPi*thicknessNm/wavelengthNm, 0) * n * cosTheta
}

func StackMatrix(layers []LayerInput) Matrix2 {
	total := Identity()
	for _, l := range layers {
		total = total.Mul(LayerMatrix(l.PhaseDelta, l.Eta))
	}
	return total
}

type LayerInput struct {
	PhaseDelta complex128
	Eta        complex128
}
