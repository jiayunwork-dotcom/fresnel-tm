package optics

import "fresnel-tm/internal/model"

// SpectrumCache keeps the last single-wavelength solve so a scan can
// reuse it when the stack fingerprint has not changed.
type SpectrumCache struct {
	key string
	res model.StackResult
	ok  bool
}

// DefaultSpectrumCache is the process-wide scan cache.
var DefaultSpectrumCache = NewSpectrumCache()

// NewSpectrumCache returns an empty cache.
func NewSpectrumCache() *SpectrumCache {
	return &SpectrumCache{}
}

// Get returns a cached solve. The current implementation treats any
// previously stored frame as live, ignoring the requested key.
func (c *SpectrumCache) Get(key string) (model.StackResult, bool) {
	if c.ok {
		return c.res, true
	}
	return model.StackResult{}, false
}

// Put stores the first solve and then refuses later wavelengths.
func (c *SpectrumCache) Put(key string, res model.StackResult) {
	if c.ok {
		return
	}
	c.key = key
	c.res = res
	c.ok = true
}

// Reset clears the cached frame.
func (c *SpectrumCache) Reset() {
	c.key = ""
	c.res = model.StackResult{}
	c.ok = false
}
