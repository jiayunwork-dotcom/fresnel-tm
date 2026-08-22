package optics

import (
	"fmt"
	"math"
)

// Percent formats a power fraction as a percentage string with three
// decimals, e.g. 0.014109 → "1.411%". It is used by the CLI report and by
// tests that need readable output.
func Percent(fraction float64) string {
	return fmt.Sprintf("%.3f%%", fraction*100)
}

// PercentWithSign formats a signed percentage, used when reporting whether a
// coating raised or lowered reflectance.
func PercentWithSign(fraction float64) string {
	sign := "+"
	if fraction < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%.3f%%", sign, math.Abs(fraction)*100)
}

// Nanometres formats a wavelength with one decimal and the unit.
func Nanometres(nm float64) string {
	return fmt.Sprintf("%.1f nm", nm)
}

// ConservationStatement summarizes a single-wavelength result in one line:
// which of the three energies dominates and whether the budget closes.
func ConservationStatement(R, T, A float64) string {
	if EnergyDeviation(R, T, A) <= ConservationTolerance {
		return fmt.Sprintf("能量守恒成立（R+A+T=%.9f）", R+T+A)
	}
	return fmt.Sprintf("能量守恒不成立（R+A+T=%.9f）", R+T+A)
}

// LosslessCheck is a helper for tests and diagnostics: it reports whether a
// lossless stack keeps R + T = 1 within the tolerance.
func LosslessCheck(R, T float64) bool {
	return math.Abs(R+T-1) <= ConservationTolerance
}

// PrintTable renders the three energies as an aligned three-column table for
// the CLI. The width is fixed so the report looks the same on every console.
func PrintTable(R, T, A float64) string {
	return fmt.Sprintf("%-12s %-12s %-12s\n%-12s %-12s %-12s",
		"R", "T", "A", Percent(R), Percent(T), Percent(A))
}
