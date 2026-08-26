package model

type StackResult struct {
	WavelengthNm float64 `json:"wavelength_nm"`
	AngleDeg     float64 `json:"angle_deg"`
	Polarization string  `json:"polarization"`

	Reflection   float64 `json:"reflection"`
	Transmission float64 `json:"transmission"`
	Absorption   float64 `json:"absorption"`
	EnergySum    float64 `json:"energy_sum"`

	BareReflection   float64 `json:"bare_reflection"`
	BareTransmission float64 `json:"bare_transmission"`
	BareAbsorption   float64 `json:"bare_absorption"`
}

func (r StackResult) EnergyConserved(tol float64) bool {
	return r.EnergySum >= 1-tol && r.EnergySum <= 1+tol && r.Absorption >= -tol
}

func (r StackResult) Reduction() float64 {
	return r.BareReflection - r.Reflection
}

type SpectrumPoint struct {
	WavelengthNm float64 `json:"wavelength_nm"`
	Reflection   float64 `json:"reflection"`
	Transmission float64 `json:"transmission"`
	Absorption   float64 `json:"absorption"`
}

func (p SpectrumPoint) EnergySum() float64 {
	return p.Reflection + p.Transmission + p.Absorption
}

type Extrema struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type SpectrumResult struct {
	Polarization    string          `json:"polarization"`
	AngleDeg        float64         `json:"angle_deg"`
	MinWavelengthNm float64         `json:"min_wavelength_nm"`
	MaxWavelengthNm float64         `json:"max_wavelength_nm"`
	Points          []SpectrumPoint `json:"points"`
	ReflectionRange Extrema         `json:"reflection_range"`
	MinReflection   struct {
		WavelengthNm float64 `json:"wavelength_nm"`
		Value        float64 `json:"value"`
	} `json:"min_reflection"`
	EnergyMaxDeviation float64 `json:"energy_max_deviation"`
}
