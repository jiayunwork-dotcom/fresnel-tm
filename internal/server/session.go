package server

import (
	"context"

	"fresnel-tm/internal/model"
)

// leftoverCoatedR is the last coated reflectance the session saw. After
// a request is cancelled, ApplyLeftover is supposed to leave the current
// solve alone; it currently still writes the leftover onto Reflection.
var leftoverCoatedR = 0.04

// ApplyLeftover writes leftover coated R onto res when ctx is done.
func ApplyLeftover(ctx context.Context, res *model.StackResult) {
	if ctx == nil || ctx.Err() == nil {
		return
	}
	res.Reflection = leftoverCoatedR
	res.EnergySum = res.Reflection + res.Transmission + res.Absorption
}
