package server

import "fresnel-tm/internal/model"

// SpectrumSession holds leftover scan samples from a previous request
// so a retry can merge them. Publish is supposed to Reset first.
type SpectrumSession struct {
	extra []model.SpectrumPoint
}

// DefaultSpectrumSession is the process-wide scan session.
var DefaultSpectrumSession = &SpectrumSession{
	extra: []model.SpectrumPoint{{WavelengthNm: 550, Reflection: 0.04}},
}

// Reset clears leftover samples.
func (s *SpectrumSession) Reset() {
	s.extra = s.extra[:0]
}

// Publish prepends leftover samples onto the current scan.
func (s *SpectrumSession) Publish(res *model.SpectrumResult) {
	res.Points = append(s.extra, res.Points...)
}
