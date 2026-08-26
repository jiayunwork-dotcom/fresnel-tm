package optics

import "math"

const ConservationTolerance = 1e-9

func EnergySum(R, T, A float64) float64 {
	return R + T + A
}

func ResidualAbsorption(R, T float64) float64 {
	return 1 - R - T
}

func EnergyDeviation(R, T, A float64) float64 {
	d := EnergySum(R, T, A) - 1
	if d < 0 {
		return -d
	}
	return d
}

func PhysicallyBounded(R, T, A float64, tol float64) bool {
	if R < -tol || T < -tol || A < -tol {
		return false
	}
	if R > 1+tol || T > 1+tol || A > 1+tol {
		return false
	}
	return EnergyDeviation(R, T, A) <= tol
}

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
