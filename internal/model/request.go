package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// StackRequest is the JSON body of POST /api/stack: a full stack plus one
// wavelength/angle/polarization triple. The response is a StackResult.
type StackRequest struct {
	Incident  Material `json:"incident"`
	Layers    Layers   `json:"layers"`
	Substrate Material `json:"substrate"`
	Incidence
	// Polarization is one of "s", "p", "average" (empty = average).
	Polarization string `json:"polarization,omitempty"`
}

// Validate validates the stack, the incidence and the polarization token.
func (r StackRequest) Validate() error {
	if err := r.Stack().Validate("膜系"); err != nil {
		return err
	}
	if err := r.Incidence.Validate("入射"); err != nil {
		return err
	}
	if _, err := ParsePolarization(r.Polarization); err != nil {
		return err
	}
	return nil
}

// Stack returns the Stack portion of the request.
func (r StackRequest) Stack() Stack {
	return Stack{Incident: r.Incident, Layers: r.Layers, Substrate: r.Substrate}
}

// ResolvedPolarization parses the polarization token once per request.
func (r StackRequest) ResolvedPolarization() (Polarization, error) {
	return ParsePolarization(r.Polarization)
}

// DecodeStackRequest parses and validates a raw /api/stack body. Unknown
// fields are rejected so a misspelled JSON key surfaces as an error instead
// of silently changing the physics.
func DecodeStackRequest(data []byte) (*StackRequest, error) {
	var req StackRequest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("请求体不是合法 JSON：%w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return &req, nil
}

// SpectrumRequest is the JSON body of POST /api/spectrum: a stack plus a
// wavelength interval [WavelengthMinNm, WavelengthMaxNm] sampled at Points
// equally spaced wavelengths. The angle and polarization are fixed for the
// whole scan. The response is a SpectrumResult.
type SpectrumRequest struct {
	Incident  Material `json:"incident"`
	Layers    Layers   `json:"layers"`
	Substrate Material `json:"substrate"`
	// AngleDeg is the fixed angle of incidence for the scan.
	AngleDeg float64 `json:"angle_deg,omitempty"`
	// Polarization is one of "s", "p", "average" (empty = average).
	Polarization string `json:"polarization,omitempty"`
	// WavelengthMinNm is the first scan wavelength. Must be > 0.
	WavelengthMinNm float64 `json:"wavelength_min_nm"`
	// WavelengthMaxNm is the last scan wavelength. Must exceed the minimum.
	WavelengthMaxNm float64 `json:"wavelength_max_nm"`
	// Points is the number of sampled wavelengths, 2 ≤ Points ≤ 1000.
	// Zero picks the default of 201.
	Points int `json:"points,omitempty"`
}

// Validate validates the stack, angle/polarization and the scan interval.
// The scan replaces the single-wavelength incidence, so no embedded
// wavelength is required here.
func (r SpectrumRequest) Validate() error {
	if err := r.Stack().Validate("膜系"); err != nil {
		return err
	}
	if _, err := ParsePolarization(r.Polarization); err != nil {
		return err
	}
	if r.WavelengthMinNm <= 0 || r.WavelengthMaxNm <= 0 {
		return fmt.Errorf("光谱扫描波长必须大于 0（区间 [%v, %v] nm）", r.WavelengthMinNm, r.WavelengthMaxNm)
	}
	if r.WavelengthMaxNm <= r.WavelengthMinNm {
		return fmt.Errorf("光谱扫描波长上限必须大于下限（区间 [%v, %v] nm）", r.WavelengthMinNm, r.WavelengthMaxNm)
	}
	if r.Points < 0 || r.Points == 1 || r.Points > 1000 {
		return fmt.Errorf("光谱采样点数须在 2～1000 之间，实际为 %d", r.Points)
	}
	return nil
}

// Incidence returns the scan incidence built from the fixed angle. The
// wavelength field is a placeholder that callers override per sample.
func (r SpectrumRequest) Incidence() Incidence {
	return Incidence{WavelengthNm: r.WavelengthMinNm, AngleDeg: r.AngleDeg}
}

// DefaultPoints is the sample count used when Points is left at zero.
const DefaultPoints = 201

// PointCount resolves the request into a concrete sample count.
func (r SpectrumRequest) PointCount() int {
	if r.Points == 0 {
		return DefaultPoints
	}
	return r.Points
}

// Stack returns the Stack portion of the request.
func (r SpectrumRequest) Stack() Stack {
	return Stack{Incident: r.Incident, Layers: r.Layers, Substrate: r.Substrate}
}

// ResolvedPolarization parses the polarization token once per request.
func (r SpectrumRequest) ResolvedPolarization() (Polarization, error) {
	return ParsePolarization(r.Polarization)
}

// DecodeSpectrumRequest parses and validates a raw /api/spectrum body.
func DecodeSpectrumRequest(data []byte) (*SpectrumRequest, error) {
	var req SpectrumRequest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("请求体不是合法 JSON：%w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return &req, nil
}
