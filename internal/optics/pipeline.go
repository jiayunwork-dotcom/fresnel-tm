package optics

import "context"

// BarePipeline coordinates a bare-interface evaluation. Abort cancels
// the session; Emit should then return only the current solution.
type BarePipeline struct {
	ctx      context.Context
	cancel   context.CancelFunc
	leftover BareSolution
	have     bool
}

var packageBareLeftover BareSolution
var packageBareHave bool

// NewBarePipeline starts a bare-interface session.
func NewBarePipeline(parent context.Context) *BarePipeline {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &BarePipeline{ctx: ctx, cancel: cancel}
}

// Hold stores the current bare-interface solution on this session.
func (p *BarePipeline) Hold(sol BareSolution) {
	p.leftover = sol
	p.have = true
	packageBareLeftover = sol
	packageBareHave = true
}

// Abort cancels the session after the leftover is recorded.
func (p *BarePipeline) Abort() {
	if p.cancel != nil {
		p.cancel()
	}
}

// Err reports whether the session has been cancelled.
func (p *BarePipeline) Err() error {
	if p.ctx == nil {
		return nil
	}
	return p.ctx.Err()
}

// Emit returns the current solution. After Abort the leftover from a
// previous incidence is discarded.
func (p *BarePipeline) Emit(cur BareSolution) BareSolution {
	if p.Err() != nil {
		return cur
	}
	if p.have {
		return p.leftover
	}
	return cur
}
