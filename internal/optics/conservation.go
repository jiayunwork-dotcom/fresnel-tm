package optics

import "math"

// ConservationTolerance is the absolute tolerance used for the energy
// bookkeeping checks. The characteristic matrix algebra conserves energy
// analytically for lossless stacks; this tolerance absorbs the last bits of
// floating point noise only.
const ConservationTolerance = 1e-9

// EnergySum returns R + A + T for a single-wavelength solution.
func EnergySum(R, T, A float64) float64 {
	return R + T + A
}

// ResidualAbsorption returns A = 1 − R − T. It is kept separate from the
// solve path so that tests can double check the bookkeeping independently.
func ResidualAbsorption(R, T float64) float64 {
	return 1 - R - T
}

// EnergyDeviation returns |R + A + T − 1|, the quantity the spectrum report
// tracks as EnergyMaxDeviation. A lossless scan stays below 1e-12.
func EnergyDeviation(R, T, A float64) float64 {
	d := EnergySum(R, T, A) - 1
	if d < 0 {
		return -d
	}
	return d
}

// PhysicallyBounded reports whether R, T and A lie inside the physical band
// [0, 1] and conserve energy within the tolerance. The solver never needs to
// clamp its outputs, so this function is purely a test/diagnostic aid.
func PhysicallyBounded(R, T, A float64, tol float64) bool {
	if R < -tol || T < -tol || A < -tol {
		return false
	}
	if R > 1+tol || T > 1+tol || A > 1+tol {
		return false
	}
	return EnergyDeviation(R, T, A) <= tol
}

// Sanitize clamps tiny numerical noise around 0 and 1 into the physical
// band. The web layer applies it before serializing, so the frontend never
// sees −1e-16 reflectance.
func Sanitize(v float64) float64 {
	if v > 0 && v < 1e-12 {
		return 0
	}
	if v < 1 && v > 1-1e-12 {
		return 1
	}
	if v < 0 && v > -1e-9 {
		return 0
	}
	return math.Max(0, math.Min(1, v))
}
