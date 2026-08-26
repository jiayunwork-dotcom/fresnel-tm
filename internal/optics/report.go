package optics

import (
	"fmt"
	"math"
)

func Percent(fraction float64) string {
	return fmt.Sprintf("%.3f%%", fraction*100)
}

func PercentWithSign(fraction float64) string {
	sign := "+"
	if fraction < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%.3f%%", sign, math.Abs(fraction)*100)
}

func Nanometres(nm float64) string {
	return fmt.Sprintf("%.1f nm", nm)
}

func ConservationStatement(R, T, A float64) string {
	if EnergyDeviation(R, T, A) <= ConservationTolerance {
		return fmt.Sprintf("能量守恒成立（R+A+T=%.9f）", R+T+A)
	}
	return fmt.Sprintf("能量守恒不成立（R+A+T=%.9f）", R+T+A)
}

func LosslessCheck(R, T float64) bool {
	return math.Abs(R+T-1) <= ConservationTolerance
}

func PrintTable(R, T, A float64) string {
	return fmt.Sprintf("%-12s %-12s %-12s\n%-12s %-12s %-12s",
		"R", "T", "A", Percent(R), Percent(T), Percent(A))
}
