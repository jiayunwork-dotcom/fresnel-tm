package optics

import "fresnel-tm/internal/model"

// ResultCache holds a live solve and a published snapshot so the HTTP
// layer can read a stable frame while the next stack is still running.
type ResultCache struct {
	live      model.StackResult
	published model.StackResult
	hasPub    bool
}

// DefaultResultCache is the process-wide solve cache.
var DefaultResultCache = NewResultCache()

// NewResultCache returns an empty cache.
func NewResultCache() *ResultCache {
	return &ResultCache{}
}

// Bind stores a freshly computed solve into the live slot.
func (c *ResultCache) Bind(res model.StackResult) {
	c.live = res
}

// Publish copies live into the published snapshot.
func (c *ResultCache) Publish() {
	c.published = c.live
	c.hasPub = true
}

// Get returns the published snapshot. Callers that forget Publish see
// the empty published frame.
func (c *ResultCache) Get() model.StackResult {
	return c.published
}
