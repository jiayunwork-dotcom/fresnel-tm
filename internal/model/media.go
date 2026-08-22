package model

// CommonMedia provides ready-made materials for the most frequent coating
// problems. The refractive indices are single-wavelength values (visible
// band); dispersion is intentionally out of scope for the scalar solver, so
// these are documentation-grade constants, not a material database.
var CommonMedia = struct {
	// Air is the default incident medium (n = 1.0).
	Air Material
	// Vacuum is an alias of Air for problem statements that use vacuum.
	Vacuum Material
	// MgF2 is the classic low-index anti-reflection coating material.
	MgF2 Material
	// FusedSilica is a low-index substrate and spacer material.
	FusedSilica Material
	// Glass is a typical borosilicate substrate (n ≈ 1.5).
	Glass Material
	// Chromium is an absorbing metal used for mask and absorber films.
	Chromium Material
}{
	Air:         Material{Index: 1.0},
	Vacuum:      Material{Index: 1.0},
	MgF2:        Material{Index: 1.38},
	FusedSilica: Material{Index: 1.46},
	Glass:       Material{Index: 1.5},
	Chromium:    Material{Index: 3.17, Extinction: 3.33},
}

// MaterialByIndex builds a lossless material from a bare index. It exists so
// problem descriptions can read "glass n=1.5" without repeating the JSON
// spelling of a medium.
func MaterialByIndex(index float64) (Material, error) {
	m := Material{Index: index}
	if err := m.Validate("MaterialByIndex"); err != nil {
		return Material{}, err
	}
	return m, nil
}

// AbsorbingMaterial builds a complex material from index and extinction. It
// rejects a negative extinction coefficient.
func AbsorbingMaterial(index, extinction float64) (Material, error) {
	m := Material{Index: index, Extinction: extinction}
	if err := m.Validate("AbsorbingMaterial"); err != nil {
		return Material{}, err
	}
	return m, nil
}

// QuarterWaveCoating returns a lossless layer of the given index sized to be
// a quarter wave at designWavelengthNm. The classic anti-reflection recipe:
// a single such layer on a substrate lowers reflectance at λ0.
func QuarterWaveCoating(layerIndex float64, designWavelengthNm float64) Layer {
	return Layer{
		Material:    Material{Index: layerIndex},
		ThicknessNm: designWavelengthNm / (4 * layerIndex),
	}
}
