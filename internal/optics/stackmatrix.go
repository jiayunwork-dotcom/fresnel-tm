package optics

import (
	"fmt"

	"fresnel-tm/internal/model"
)

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
