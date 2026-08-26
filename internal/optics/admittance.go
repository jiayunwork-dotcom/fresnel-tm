package optics

import (
	"fresnel-tm/internal/model"
)

func Admittance(pol model.Polarization, n, cosTheta complex128) complex128 {
	if pol == model.PolP {
		return n / cosTheta
	}
	return n * cosTheta
}

func RealPart(eta complex128) float64 {
	return real(eta)
}

func ImagPartAbs(eta complex128) float64 {
	i := imag(eta)
	if i < 0 {
		return -i
	}
	return i
}

func AdmittancePair(pol model.Polarization, nInc complex128, cosInc complex128, nSub complex128, cosSub complex128) (eta0, etaSub complex128) {
	eta0 = Admittance(pol, nInc, cosInc)
	etaSub = Admittance(pol, nSub, cosSub)
	return eta0, etaSub
}
