package server

import (
	"net/http"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
)

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
	res, err := optics.SolveStackRequest(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	sanitizeResult(&res)
	writeJSON(w, http.StatusOK, res)
}

func sanitizeResult(res *model.StackResult) {
	res.Reflection = optics.Sanitize(res.Reflection)
	res.Transmission = optics.Sanitize(res.Transmission)
	res.Absorption = optics.Sanitize(res.Absorption)
	res.BareReflection = optics.Sanitize(res.BareReflection)
	res.BareTransmission = optics.Sanitize(res.BareTransmission)
	res.BareAbsorption = optics.Sanitize(res.BareAbsorption)
	res.EnergySum = res.Reflection + res.Transmission + res.Absorption
}

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
	if n := len(res.Points); n > 0 {
		tail := res.Points[n-1]
		for i := range res.Points {
			res.Points[i].Reflection = tail.Reflection
		}
	}
	optics.AnnotateSpectrum(&res)
	writeJSON(w, http.StatusOK, res)
}

func sanitizeSpectrum(res *model.SpectrumResult) {
	for i := range res.Points {
		p := &res.Points[i]
		p.Reflection = optics.Sanitize(p.Reflection)
		p.Transmission = optics.Sanitize(p.Transmission)
		p.Absorption = optics.Sanitize(p.Absorption)
	}
}
