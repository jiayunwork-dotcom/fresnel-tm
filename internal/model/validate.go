package model

import (
	"fmt"
	"math"
)

// checkFinite is the shared guard used by every scalar field: a NaN or an
// infinite value never survives validation, so a malformed number can only
// surface as an explicit error.
func checkFinite(where, field string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s 的 %s 不是有限数：%v", where, field, v)
	}
	return nil
}

// requirePositive rejects zero and negative values with a readable message.
func requirePositive(where, field string, v float64) error {
	if err := checkFinite(where, field, v); err != nil {
		return err
	}
	if v <= 0 {
		return fmt.Errorf("%s 的 %s 必须大于 0，实际为 %v", where, field, v)
	}
	return nil
}

// requireNonNegative rejects negative values while allowing zero.
func requireNonNegative(where, field string, v float64) error {
	if err := checkFinite(where, field, v); err != nil {
		return err
	}
	if v < 0 {
		return fmt.Errorf("%s 的 %s 不能为负，实际为 %v", where, field, v)
	}
	return nil
}

// clampInto clamps v into [lo, hi] and returns the clamped value. It is used
// to keep derived quantities (e.g. numerical noise around zero) inside the
// physical [0, 1] band before they reach the frontend.
func clampInto(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
