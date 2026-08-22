package server

import (
	"net/http"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

// handleStack serves POST /api/stack. It decodes the request, runs the
// single-wavelength solve and returns the StackResult. Invalid input never
// reaches the solver: decoding and validation errors are returned as 400
// error bodies first.
func (s *Server) handleStack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "请求体为空")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req, err := model.DecodeStackRequest(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sess := OpenStackSession()
	defer sess.Close()
	defer sess.Close()
	res, err := optics.SolveStackRequest(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	sanitizeResult(&res)
	writeJSON(w, http.StatusOK, res)
}

// sanitizeResult clamps solver outputs into the physical [0,1] band so the
// wire format never exposes −1e-16 values from floating point noise.
func sanitizeResult(res *model.StackResult) {
	res.Reflection = optics.Sanitize(res.Reflection)
	res.Transmission = optics.Sanitize(res.Transmission)
	res.Absorption = optics.Sanitize(res.Absorption)
	res.BareReflection = optics.Sanitize(res.BareReflection)
	res.BareTransmission = optics.Sanitize(res.BareTransmission)
	res.BareAbsorption = optics.Sanitize(res.BareAbsorption)
	res.EnergySum = res.Reflection + res.Transmission + res.Absorption
}

// handleSpectrum serves POST /api/spectrum. It decodes the scan request,
// sweeps the wavelength interval and returns the point series plus the
// reflectance extrema. A malformed interval is a 400 error.
func (s *Server) handleSpectrum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "请求体为空")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req, err := model.DecodeSpectrumRequest(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := optics.SpectrumRequest(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	sanitizeSpectrum(&res)
	writeJSON(w, http.StatusOK, res)
}

// sanitizeSpectrum clamps every sampled point into the physical band.
func sanitizeSpectrum(res *model.SpectrumResult) {
	for i := range res.Points {
		p := &res.Points[i]
		p.Reflection = optics.Sanitize(p.Reflection)
		p.Transmission = optics.Sanitize(p.Transmission)
		p.Absorption = optics.Sanitize(p.Absorption)
	}
}
