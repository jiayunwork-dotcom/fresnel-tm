package optics

// Denominator is the shared expression
//
//	D = η0·M11 + η0·ηs·M12 + M21 + ηs·M22
//
// that both amplitude coefficients divide by. Keeping one implementation
// guarantees r and t use the same phase reference.
func Denominator(eta0, etaSub complex128, m Matrix2) complex128 {
	return eta0*m.A + eta0*etaSub*m.B + m.C + etaSub*m.D
}

// ReflectionCoefficient returns the complex amplitude reflection coefficient
// r of the whole system:
//
//	r = (η0·M11 + η0·ηs·M12 − M21 − ηs·M22) / D
//
// For a bare interface (M = I) this collapses to the Fresnel coefficient
// (η0 − ηs)/(η0 + ηs).
func ReflectionCoefficient(eta0, etaSub complex128, m Matrix2) complex128 {
	N := eta0*m.A + eta0*etaSub*m.B - m.C - etaSub*m.D
	return N / Denominator(eta0, etaSub, m)
}

// TransmissionCoefficient returns the complex amplitude transmission
// coefficient t = 2·η0 / D. Its power fraction is scaled by the admittance
// ratio in Transmittance.
func TransmissionCoefficient(eta0, etaSub complex128, m Matrix2) complex128 {
	return 2 * eta0 / Denominator(eta0, etaSub, m)
}

// Reflectance returns the power reflectance R = |r|².
func Reflectance(r complex128) float64 {
	return Abs2(r)
}

// Transmittance returns the power transmittance
//
//	T = (Re ηs / Re η0) · |t|²
//
// The admittance ratio is the "angle factor": for oblique incidence the
// transmitted power differs from |t|² by exactly this ratio, and omitting it
// breaks R + T = 1. Re η0 is positive because the incident medium is
// lossless.
func Transmittance(eta0, etaSub complex128, t complex128) float64 {
	ratio := RealPart(etaSub) / RealPart(eta0)
	return ratio * Abs2(t)
}
