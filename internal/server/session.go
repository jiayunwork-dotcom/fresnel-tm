package server

import (
	"context"

	"fresnel-tm/internal/model"
)

// leftoverCoatedR tracks the last coated reflectance the session saw.
var leftoverCoatedR = 0.04

// ApplyLeftover records the current coated R. A cancelled request still
// keeps that freshly published result; leftover glass R is not written back.
func ApplyLeftover(ctx context.Context, res *model.StackResult) {
	if res == nil {
		return
	}
	leftoverCoatedR = res.Reflection
	if ctx == nil || ctx.Err() == nil {
		return
	}
}
