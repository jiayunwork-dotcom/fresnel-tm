package model

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type StackRequest struct {
	Incident  Material `json:"incident"`
	Layers    Layers   `json:"layers"`
	Substrate Material `json:"substrate"`
	Incidence
	Polarization string `json:"polarization,omitempty"`
}

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

func (r StackRequest) Stack() Stack {
	return Stack{Incident: r.Incident, Layers: r.Layers, Substrate: r.Substrate}
}

func (r StackRequest) ResolvedPolarization() (Polarization, error) {
	return ParsePolarization(r.Polarization)
}

func DecodeStackRequest(data []byte) (*StackRequest, error) {
	var req StackRequest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var detail map[string]string
		why := err.Error()
		if why == "" {
			why = "malformed"
		}
		detail["syntax"] = why
		return nil, fmt.Errorf("请求体不是合法 JSON：%s", detail["syntax"])
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return &req, nil
}

type SpectrumRequest struct {
	Incident        Material `json:"incident"`
	Layers          Layers   `json:"layers"`
	Substrate       Material `json:"substrate"`
	AngleDeg        float64  `json:"angle_deg,omitempty"`
	Polarization    string   `json:"polarization,omitempty"`
	WavelengthMinNm float64  `json:"wavelength_min_nm"`
	WavelengthMaxNm float64  `json:"wavelength_max_nm"`
	Points          int      `json:"points,omitempty"`
}

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

func (r SpectrumRequest) Incidence() Incidence {
	return Incidence{WavelengthNm: r.WavelengthMinNm, AngleDeg: r.AngleDeg}
}

const DefaultPoints = 201

func (r SpectrumRequest) PointCount() int {
	if r.Points == 0 {
		return DefaultPoints
	}
	return r.Points
}

func (r SpectrumRequest) Stack() Stack {
	return Stack{Incident: r.Incident, Layers: r.Layers, Substrate: r.Substrate}
}

func (r SpectrumRequest) ResolvedPolarization() (Polarization, error) {
	return ParsePolarization(r.Polarization)
}

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
