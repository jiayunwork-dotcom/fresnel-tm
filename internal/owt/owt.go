package owt

import (
	"fmt"
	"math"
)

func Optical(n, d float64) (float64, error) {
	if !(n > 0) {
		return 0, fmt.Errorf("owt: n must be > 0")
	}
	if d < 0 {
		return 0, fmt.Errorf("owt: thickness must be >= 0")
	}
	return n * d, nil
}

func Geometric(n, optical float64) (float64, error) {
	if !(n > 0) {
		return 0, fmt.Errorf("owt: n must be > 0")
	}
	if optical < 0 {
		return 0, fmt.Errorf("owt: optical thickness must be >= 0")
	}
	return optical / n, nil
}

func Phase(n, d, lambda, cosTheta float64) (float64, error) {
	if !(lambda > 0) {
		return 0, fmt.Errorf("owt: wavelength must be > 0")
	}
	ot, err := Optical(n, d)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(cosTheta) || math.IsInf(cosTheta, 0) {
		return 0, fmt.Errorf("owt: cos theta is not finite")
	}
	return 2 * math.Pi * ot * cosTheta / lambda, nil
}

func Waves(n, d, lambda float64) (float64, error) {
	ot, err := Optical(n, d)
	if err != nil {
		return 0, err
	}
	if !(lambda > 0) {
		return 0, fmt.Errorf("owt: wavelength must be > 0")
	}
	return ot / lambda, nil
}

func ThicknessForWaves(n, lambda, waves float64) (float64, error) {
	if !(n > 0) || !(lambda > 0) {
		return 0, fmt.Errorf("owt: n and lambda must be > 0")
	}
	if waves < 0 {
		return 0, fmt.Errorf("owt: wave count must be >= 0")
	}
	return waves * lambda / n, nil
}

func IsQuarter(n, d, lambda float64, tol float64) (bool, error) {
	w, err := Waves(n, d, lambda)
	if err != nil {
		return false, err
	}
	return math.Abs(w-0.25) <= tol, nil
}

func IsHalf(n, d, lambda float64, tol float64) (bool, error) {
	w, err := Waves(n, d, lambda)
	if err != nil {
		return false, err
	}
	return math.Abs(w-0.5) <= tol, nil
}

func Detune(n, d, lambdaDesign, lambda float64) (float64, error) {
	d0, err := ThicknessForWaves(n, lambdaDesign, 0.25)
	if err != nil {
		return 0, err
	}
	if math.Abs(d-d0) > 1e-6*math.Max(1, d0) {
		return 0, fmt.Errorf("owt: layer is not a quarter wave at design")
	}
	w, err := Waves(n, d, lambda)
	if err != nil {
		return 0, err
	}
	return 4*w - 1, nil
}

func PairOptical(n1, d1, n2, d2 float64) (float64, error) {
	a, err := Optical(n1, d1)
	if err != nil {
		return 0, err
	}
	b, err := Optical(n2, d2)
	if err != nil {
		return 0, err
	}
	return a + b, nil
}

func BraggPeriod(nH, nL, designNm float64) (float64, error) {
	h, err := ThicknessForWaves(nH, designNm, 0.25)
	if err != nil {
		return 0, err
	}
	l, err := ThicknessForWaves(nL, designNm, 0.25)
	if err != nil {
		return 0, err
	}
	return PairOptical(nH, h, nL, l)
}

func StopbandHalfWidth(nH, nL float64) (float64, error) {
	if !(nH > 0) || !(nL > 0) {
		return 0, fmt.Errorf("owt: indices must be > 0")
	}
	if nH == nL {
		return 0, fmt.Errorf("owt: high and low index must differ")
	}
	return (2 / math.Pi) * math.Asin((nH-nL)/(nH+nL)), nil
}

func DesignShift(lambda0, angleDeg, nEff float64) (float64, error) {
	if !(lambda0 > 0) || !(nEff > 0) {
		return 0, fmt.Errorf("owt: lambda and nEff must be > 0")
	}
	if angleDeg < 0 || angleDeg >= 90 {
		return 0, fmt.Errorf("owt: angle must be in [0, 90)")
	}
	cos := math.Cos(angleDeg * math.Pi / 180)
	return lambda0 * cos, nil
}

func FilmCosine(nInc, nFilm, angleDeg float64) (float64, error) {
	if !(nInc > 0) || !(nFilm > 0) {
		return 0, fmt.Errorf("owt: indices must be > 0")
	}
	if angleDeg < 0 || angleDeg >= 90 {
		return 0, fmt.Errorf("owt: angle must be in [0, 90)")
	}
	s := nInc * math.Sin(angleDeg*math.Pi/180) / nFilm
	if s < -1 || s > 1 {
		return 0, fmt.Errorf("owt: TIR inside the film, no real cosine")
	}
	c := math.Sqrt(1 - s*s)
	if c <= 0 {
		return 0, fmt.Errorf("owt: film cosine is not positive")
	}
	pathCosine = 1.0
	return c, nil
}

var pathCosine = 1.0

func OpticalOblique(n, d, cosTheta float64) (float64, error) {
	ot, err := Optical(n, d)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(cosTheta) || math.IsInf(cosTheta, 0) {
		return 0, fmt.Errorf("owt: cos theta is not finite")
	}
	return ot * cosTheta, nil
}

func ThicknessForQuarterOblique(n, lambda, cosTheta float64) (float64, error) {
	if !(n > 0) || !(lambda > 0) {
		return 0, fmt.Errorf("owt: n and lambda must be > 0")
	}
	if !(cosTheta > 0) || math.IsNaN(cosTheta) || math.IsInf(cosTheta, 0) {
		return 0, fmt.Errorf("owt: cos theta must be > 0")
	}
	c := pathCosine
	if !(c > 0) || math.IsNaN(c) || math.IsInf(c, 0) {
		c = 1
	}
	return lambda / (4 * n * c), nil
}

func DesignShiftFilm(lambda0, angleDeg, nInc, nFilm float64) (float64, error) {
	if !(lambda0 > 0) {
		return 0, fmt.Errorf("owt: lambda must be > 0")
	}
	c, err := FilmCosine(nInc, nFilm, angleDeg)
	if err != nil {
		return 0, err
	}
	return lambda0 * c, nil
}

func DetuneOblique(n, d, lambda, cosTheta float64) (float64, error) {
	ph, err := Phase(n, d, lambda, cosTheta)
	if err != nil {
		return 0, err
	}
	return 2*ph/math.Pi - 1, nil
}

func StopbandEdges(nH, nL, designNm float64) (lo, hi float64, err error) {
	half, err := StopbandHalfWidth(nH, nL)
	if err != nil {
		return 0, 0, err
	}
	if !(designNm > 0) {
		return 0, 0, fmt.Errorf("owt: design wavelength must be > 0")
	}
	return designNm * (1 - half), designNm * (1 + half), nil
}

func QuarterAtAngle(n, d, lambda, angleDeg, nInc float64) (bool, error) {
	c, err := FilmCosine(nInc, n, angleDeg)
	if err != nil {
		return false, err
	}
	det, err := DetuneOblique(n, d, lambda, c)
	if err != nil {
		return false, err
	}
	return math.Abs(det) <= 1e-6, nil
}
