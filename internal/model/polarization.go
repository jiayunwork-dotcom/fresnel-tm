package model

import (
	"fmt"
	"strings"
)

// Polarization selects which transverse polarization is solved. The layer
// admittance η differs between the two polarizations — η = n̂·cosθ for s
// (TE, electric field perpendicular to the plane of incidence) and η = n̂/cosθ
// for p (TM, electric field parallel to the plane of incidence). Mixing the
// two admittance rules up produces wrong spectra, so the polarization is an
// explicit part of every request.
type Polarization int

const (
	// PolS is s-polarization (TE): η = n̂·cosθ.
	PolS Polarization = iota
	// PolP is p-polarization (TM): η = n̂/cosθ.
	PolP
	// PolAverage reports the unweighted mean of the s and p results. Real
	// measurements are a mixture; the mean of the two orthogonal states is
	// the simplest unpolarized estimate.
	PolAverage
)

// ParsePolarization converts a wire-format string into a Polarization.
// Accepted values are "s", "p" and "average" (or the abbreviations "avg").
// An empty string defaults to the average. Anything else is an error so that
// a typo cannot silently select the wrong physics.
func ParsePolarization(s string) (Polarization, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "average", "avg":
		return PolAverage, nil
	case "s", "te":
		return PolS, nil
	case "p", "tm":
		return PolP, nil
	}
	return PolAverage, fmt.Errorf("未知偏振态 %q（可选：s、p、average）", s)
}

// String returns the canonical wire-format name of the polarization.
func (p Polarization) String() string {
	switch p {
	case PolS:
		return "s"
	case PolP:
		return "p"
	default:
		return "average"
	}
}

// ShortName returns a one-letter label used by the frontend legend.
func (p Polarization) ShortName() string {
	switch p {
	case PolS:
		return "s"
	case PolP:
		return "p"
	default:
		return "avg"
	}
}

// Split returns the pair of elementary polarizations the value stands for.
// PolAverage splits into both s and p; the two others return themselves.
// The solver uses Split to reduce every request to elementary s/p runs.
func (p Polarization) Split() []Polarization {
	switch p {
	case PolS:
		return []Polarization{PolS}
	case PolP:
		return []Polarization{PolP}
	default:
		return []Polarization{PolS, PolP}
	}
}

// AllPolarizations enumerates the three supported values.
func AllPolarizations() []Polarization {
	return []Polarization{PolS, PolP, PolAverage}
}
