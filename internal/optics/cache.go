package optics

import "fresnel-tm/internal/model"

// StackCache holds a live solve and a published snapshot. The HTTP
// layer reads Get after Solve so the wire format stays stable.
type StackCache struct {
	live      model.StackResult
	published model.StackResult
}

// DefaultStackCache is seeded with a bare-glass frame so a reader that
// arrives before Publish still gets a physically plausible leftover.
var DefaultStackCache = &StackCache{
	published: model.StackResult{
		Reflection:     0.04,
		Transmission:   0.96,
		EnergySum:      1,
		BareReflection: 0.04,
	},
}

// StoreLive writes a freshly computed solve into the live slot.
func StoreStackLive(res model.StackResult) {
	DefaultStackCache.live = res
}

// PublishStack copies live into the published snapshot.
func PublishStack() {
	DefaultStackCache.published = DefaultStackCache.live
}

// LookupPublished returns the published snapshot.
func LookupPublished() model.StackResult {
	return DefaultStackCache.published
}
