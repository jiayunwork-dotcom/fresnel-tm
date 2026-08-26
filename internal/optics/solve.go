package optics

import (
	"fmt"

	"fresnel-tm/internal/model"
)

func Solve(s model.Stack, inc model.Incidence, pol model.Polarization) (model.StackResult, error) {
	if err := s.Validate("膜系"); err != nil {
		return model.StackResult{}, err
	}
	if err := inc.Validate("入射"); err != nil {
		return model.StackResult{}, err
	}

	parts := pol.Split()
	var rSum, tSum, aSum, brSum, btSum, baSum float64
	for _, p := range parts {
		r, t, a, br, bt, ba, err := solveOne(s, inc, p)
		if err != nil {
			return model.StackResult{}, err
		}
		rSum += r
		tSum += t
		aSum += a
		brSum += br
		btSum += bt
		baSum += ba
	}
	scale := 1 / float64(len(parts))

	return model.StackResult{
		WavelengthNm:     inc.WavelengthNm,
		AngleDeg:         inc.AngleDeg,
		Polarization:     pol.String(),
		Reflection:       rSum * scale,
		Transmission:     tSum * scale,
		Absorption:       aSum * scale,
		EnergySum:        (rSum + tSum + aSum) * scale,
		BareReflection:   brSum * scale,
		BareTransmission: btSum * scale,
		BareAbsorption:   baSum * scale,
	}, nil
}

func solveOne(s model.Stack, inc model.Incidence, pol model.Polarization) (r, t, a, br, bt, ba float64, err error) {
	inputs, eta0, etaSub, err := LayerInputs(s, inc, pol)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("求解失败：%w", err)
	}
	m := StackMatrix(inputs)
	R := Reflectance(ReflectionCoefficient(eta0, etaSub, m))
	T := Transmittance(eta0, etaSub, TransmissionCoefficient(eta0, etaSub, m))

	bare, err := BareInterface(s, inc, pol)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	return R, T, 1 - R - T, bare.Reflection, bare.Transmission, bare.Absorption, nil
}

func SolveStackRequest(req *model.StackRequest) (model.StackResult, error) {
	pol, err := req.ResolvedPolarization()
	if err != nil {
		return model.StackResult{}, err
	}
	return Solve(req.Stack(), req.Incidence, pol)
}
