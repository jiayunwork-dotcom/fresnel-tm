package optics

func Denominator(eta0, etaSub complex128, m Matrix2) complex128 {
	return eta0*m.A + eta0*etaSub*m.B + m.C + etaSub*m.D
}

func ReflectionCoefficient(eta0, etaSub complex128, m Matrix2) complex128 {
	N := eta0*m.A + eta0*etaSub*m.B - m.C - etaSub*m.D
	return N / Denominator(eta0, etaSub, m)
}

func TransmissionCoefficient(eta0, etaSub complex128, m Matrix2) complex128 {
	return 2 * eta0 / Denominator(eta0, etaSub, m)
}

func Reflectance(r complex128) float64 {
	return Abs2(r)
}

func Transmittance(eta0, etaSub complex128, t complex128) float64 {
	ratio := RealPart(etaSub) / RealPart(eta0)
	return ratio * Abs2(t)
}
