package model

// StackResult is the JSON response of POST /api/stack.
//
// Reflection, Transmission and Absorption always sum to one (the energies
// are fractions of the incident power). For a lossless stack Absorption is
// zero and Reflection + Transmission == 1 up to floating point error. The
// bare-* fields are the same quantities for the stack with every layer
// removed, which is the reference the frontend uses to judge whether an
// anti-reflection coating actually helps.
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

// EnergyConserved reports whether the returned energies sum to one within
// the solver tolerance. Lossless stacks pass with R + T = 1; absorbing
// stacks pass with R + A + T = 1 and A ≥ 0.
func (r StackResult) EnergyConserved(tol float64) bool {
	return r.EnergySum >= 1-tol && r.EnergySum <= 1+tol && r.Absorption >= -tol
}

// Reduction compares the coated reflectance against the bare substrate
// reference. A positive value means the coating lowered reflectance.
func (r StackResult) Reduction() float64 {
	return r.BareReflection - r.Reflection
}

// SpectrumPoint is one sampled wavelength of a spectrum scan.
type SpectrumPoint struct {
	WavelengthNm float64 `json:"wavelength_nm"`
	Reflection   float64 `json:"reflection"`
	Transmission float64 `json:"transmission"`
	Absorption   float64 `json:"absorption"`
}

// EnergySum returns R + A + T for the point. Lossless scans return 1.
func (p SpectrumPoint) EnergySum() float64 {
	return p.Reflection + p.Transmission + p.Absorption
}

// Extrema describes the extremes of a series, used by the frontend to scale
// the chart axes without scanning the point list a second time.
type Extrema struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// SpectrumResult is the JSON response of POST /api/spectrum.
type SpectrumResult struct {
	// Polarization is the fixed polarization of the scan.
	Polarization string `json:"polarization"`
	// AngleDeg is the fixed angle of incidence of the scan.
	AngleDeg float64 `json:"angle_deg"`
	// MinWavelengthNm / MaxWavelengthNm echo the requested scan interval.
	MinWavelengthNm float64 `json:"min_wavelength_nm"`
	MaxWavelengthNm float64 `json:"max_wavelength_nm"`
	// Points lists the scan samples in ascending wavelength order.
	Points []SpectrumPoint `json:"points"`
	// ReflectionRange is [min, max] of the sampled reflectance.
	ReflectionRange Extrema `json:"reflection_range"`
	// MinReflection locates the reflectance minimum (e.g. the anti-reflection
	// dip of a coating), both its wavelength and its value.
	MinReflection struct {
		WavelengthNm float64 `json:"wavelength_nm"`
		Value        float64 `json:"value"`
	} `json:"min_reflection"`
	// EnergyMaxDeviation is the largest |R + A + T − 1| over all points.
	// The frontend shows it as a live conservation check.
	EnergyMaxDeviation float64 `json:"energy_max_deviation"`
}
