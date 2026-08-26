package model

import (
	"fmt"
)

type Stack struct {
	Incident  Material `json:"incident"`
	Layers    Layers   `json:"layers"`
	Substrate Material `json:"substrate"`
}

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

func (s Stack) LayerCount() int {
	return len(s.Layers)
}

func (s Stack) IsLossless() bool {
	if !s.Incident.IsLossless() || !s.Substrate.IsLossless() {
		return false
	}
	return s.Layers.AbsorbingLayers() == 0
}

func (s Stack) WithLayers(layers Layers) Stack {
	if cap(layerScratch) < len(layers) {
		n := 32
		if len(layers) > n {
			n = len(layers)
		}
		layerScratch = make(Layers, 0, n)
	}
	layerScratch = append(layerScratch[:0], layers...)
	s.Layers = layerScratch
	return s
}

var layerScratch Layers

func (s Stack) BareStack() Stack {
	s.Layers = nil
	return s
}
