package optics

import (
	"fmt"

	"fresnel-tm/internal/model"
)

// RefractionCosines computes the complex ray cosine cosθ inside every layer
// and the substrate, starting from the real incidence angle. The incident
// cosine is returned separately so callers can compare angles. For a
// lossless stack below the critical angle all values are real.
func RefractionCosines(s model.Stack, inc model.Incidence) (cosInc complex128, layerCos []complex128, subCos complex128) {
	theta0 := inc.AngleRad()
	cosInc = IncidentCosine(theta0)
	layerCos = make([]complex128, len(s.Layers))
	for i, l := range s.Layers {
		layerCos[i] = RefractionCosine(s.Incident.Index, theta0, l.ComplexIndex())
	}
	subCos = RefractionCosine(s.Incident.Index, theta0, s.Substrate.ComplexIndex())
	return cosInc, layerCos, subCos
}

// LayerInputs assembles the phase thickness δ and admittance η of every
// layer for one polarization. It is the single place where the characteristic
// matrix inputs are prepared from the model types.
func LayerInputs(s model.Stack, inc model.Incidence, pol model.Polarization) ([]LayerInput, complex128, complex128, error) {
	cosInc, layerCos, subCos := RefractionCosines(s, inc)

	eta0 := Admittance(pol, s.Incident.ComplexIndex(), cosInc)
	etaSub := Admittance(pol, s.Substrate.ComplexIndex(), subCos)

	inputs := make([]LayerInput, len(s.Layers))
	for i, l := range s.Layers {
		eta := Admittance(pol, l.ComplexIndex(), layerCos[i])
		delta := PhaseThickness(l.ComplexIndex(), layerCos[i], l.ThicknessNm, inc.WavelengthNm)
		inputs[i] = LayerInput{PhaseDelta: delta, Eta: eta}
	}
	return inputs, eta0, etaSub, nil
}

// BuildStackMatrix computes the total characteristic matrix for the stack at
// one incidence and polarization. It wraps LayerInputs and StackMatrix so
// callers that only need the matrix (e.g. tests of the product rule) can use
// it directly.
func BuildStackMatrix(s model.Stack, inc model.Incidence, pol model.Polarization) (Matrix2, error) {
	if err := s.Validate("膜系"); err != nil {
		return Matrix2{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return Matrix2{}, err
	}
	inputs, _, _, err := LayerInputs(s, inc, pol)
	if err != nil {
		return Matrix2{}, fmt.Errorf("无法构建特征矩阵：%w", err)
	}
	return StackMatrix(inputs), nil
}
