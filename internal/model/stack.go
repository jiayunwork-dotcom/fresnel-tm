package model

import (
	"fmt"
)

// Stack is the complete optical system: incident medium, ordered coating
// layers and substrate. Zero layers is legal — it describes the bare
// incident/substrate interface, whose reflectance must match the single-
// interface Fresnel result.
type Stack struct {
	// Incident is the medium the light arrives from. It must be lossless
	// (extinction 0) so that the reference admittance Re(η0) is well defined.
	Incident Material `json:"incident"`
	// Layers lists the coatings in the order light passes through them.
	Layers Layers `json:"layers"`
	// Substrate is the exit half-space and may absorb (e.g. silicon).
	Substrate Material `json:"substrate"`
}

// Validate checks the whole stack. The incident medium is required to be
// lossless: an absorbing superstrate would make the reference admittance
// complex and break the R + A + T = 1 bookkeeping.
func (s Stack) Validate(where string) error {
	if err := s.Incident.Validate("入射介质"); err != nil {
		return err
	}
	if s.Incident.Absorbs() {
		return fmt.Errorf("%s: 入射介质不能有消光（吸收超包层会让参考导纳变复数）", where)
	}
	if err := s.Substrate.Validate("衬底"); err != nil {
		return err
	}
	if err := s.Layers.Validate("膜系"); err != nil {
		return err
	}
	return nil
}

// LayerCount returns the number of coating layers.
func (s Stack) LayerCount() int {
	return len(s.Layers)
}

// IsLossless reports whether every medium in the stack (including substrate)
// has a zero extinction coefficient. Lossless stacks obey R + T = 1 exactly.
func (s Stack) IsLossless() bool {
	if !s.Incident.IsLossless() || !s.Substrate.IsLossless() {
		return false
	}
	return s.Layers.AbsorbingLayers() == 0
}

// WithLayers returns a copy of the stack with a new layer list.
func (s Stack) WithLayers(layers Layers) Stack {
	s.Layers = layers
	return s
}

// BareStack returns the same incident/substrate pair with no layers at all —
// the reference system against which anti-reflection coatings are compared.
func (s Stack) BareStack() Stack {
	s.Layers = nil
	return s
}
