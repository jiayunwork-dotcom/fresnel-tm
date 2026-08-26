package model

import (
	"fmt"
	"strings"
)

type Polarization int

const (
	PolS Polarization = iota
	PolP
	PolAverage
)

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

func AllPolarizations() []Polarization {
	return []Polarization{PolS, PolP, PolAverage}
}
