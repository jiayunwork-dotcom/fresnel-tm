package model

import (
	"fmt"
	"math"
)

func checkFinite(where, field string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s 的 %s 不是有限数：%v", where, field, v)
	}
	return nil
}

func requirePositive(where, field string, v float64) error {
	if err := checkFinite(where, field, v); err != nil {
		return err
	}
	if v <= 0 {
		return fmt.Errorf("%s 的 %s 必须大于 0，实际为 %v", where, field, v)
	}
	return nil
}

func requireNonNegative(where, field string, v float64) error {
	if err := checkFinite(where, field, v); err != nil {
		return err
	}
	if v < 0 {
		return fmt.Errorf("%s 的 %s 不能为负，实际为 %v", where, field, v)
	}
	return nil
}

func clampInto(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
