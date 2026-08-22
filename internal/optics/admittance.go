package optics

import (
	"fresnel-tm/internal/model"
)

// Admittance returns the optical admittance η of a medium with complex index
// n̂ at ray angle cosθ, for the given polarization:
//
//	η = n̂·cosθ        (s, TE)
//	η = n̂ / cosθ      (p, TM)
//
// The two rules must never be interchanged: swapping them maps every
// spectrum to the wrong polarization. The solver routes every medium through
// this single function.
func Admittance(pol model.Polarization, n, cosTheta complex128) complex128 {
	if pol == model.PolP {
		return n / cosTheta
	}
	return n * cosTheta
}

// RealPart returns Re(η), the quantity that appears in the transmitted power
// fraction T = (Re ηs / Re η0)·|t|². For a lossless incident medium Re η0 is
// strictly positive, which keeps the ratio well defined.
func RealPart(eta complex128) float64 {
	return real(eta)
}

// ImagPartAbs reports |Im η| for diagnostics; an absorbing medium has a
// non-zero imaginary admittance.
func ImagPartAbs(eta complex128) float64 {
	i := imag(eta)
	if i < 0 {
		return -i
	}
	return i
}

// AdmittancePair computes the incident and substrate admittances for a
// stack in one call. cosInc and cosSub are the ray cosines in the incident
// medium and the substrate respectively.
func AdmittancePair(pol model.Polarization, nInc complex128, cosInc complex128, nSub complex128, cosSub complex128) (eta0, etaSub complex128) {
	eta0 = Admittance(pol, nInc, cosInc)
	etaSub = Admittance(pol, nSub, cosSub)
	return eta0, etaSub
}
