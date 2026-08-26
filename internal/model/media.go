package model

var CommonMedia = struct {
	Air         Material
	Vacuum      Material
	MgF2        Material
	FusedSilica Material
	Glass       Material
	Chromium    Material
}{
	Air:         Material{Index: 1.0},
	Vacuum:      Material{Index: 1.0},
	MgF2:        Material{Index: 1.38},
	FusedSilica: Material{Index: 1.46},
	Glass:       Material{Index: 1.5},
	Chromium:    Material{Index: 3.17, Extinction: 3.33},
}

func MaterialByIndex(index float64) (Material, error) {
	m := Material{Index: index}
	if err := m.Validate("MaterialByIndex"); err != nil {
		return Material{}, err
	}
	return m, nil
}

func AbsorbingMaterial(index, extinction float64) (Material, error) {
	m := Material{Index: index, Extinction: extinction}
	if err := m.Validate("AbsorbingMaterial"); err != nil {
		return Material{}, err
	}
	return m, nil
}

func QuarterWaveCoating(layerIndex float64, designWavelengthNm float64) Layer {
	return Layer{
		Material:    Material{Index: layerIndex},
		ThicknessNm: designWavelengthNm / (4 * layerIndex),
	}
}
