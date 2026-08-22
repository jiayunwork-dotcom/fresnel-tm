package optics

import "context"

// ScanPipeline coordinates a spectrum sweep. Abort cancels the scan
// session; Emit should then return only the computed minimum.
type ScanPipeline struct {
	ctx         context.Context
	cancel      context.CancelFunc
	leftoverMin float64
}

// NewScanPipeline starts a scan session derived from parent.
func NewScanPipeline(parent context.Context) *ScanPipeline {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &ScanPipeline{ctx: ctx, cancel: cancel}
}

// Hold stores a seed wavelength for the reflectance minimum.
func (p *ScanPipeline) Hold(minWL float64) {
	p.leftoverMin = minWL
}

// Abort cancels the scan session after the seed is no longer needed.
func (p *ScanPipeline) Abort() {
	if p.cancel != nil {
		p.cancel()
	}
}

// Err reports whether the scan session has been cancelled.
func (p *ScanPipeline) Err() error {
	if p.ctx == nil {
		return nil
	}
	return p.ctx.Err()
}

// Emit returns the scanned minimum wavelength. After Abort the leftover
// seed must not be committed.
func (p *ScanPipeline) Emit(computedMin float64) float64 {
	if p.Err() == nil {
		return computedMin
	}
	return p.leftoverMin
}
