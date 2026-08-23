package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// testServer builds a server over a minimal embedded filesystem with one web
// page and one example, and returns its httptest server.
func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	assets := fstest.MapFS{
		"web/index.html":              {Data: []byte("<html>fresnel-tm console</html>")},
		"web/app.js":                  {Data: []byte("// console script")},
		"example/ar-quarter.json":     {Data: []byte(arQuarterJSON)},
		"example/absorber-stack.json": {Data: []byte(absorberJSON)},
	}
	ts := httptest.NewServer(NewServer(assets).Handler())
	t.Cleanup(ts.Close)
	return ts
}

const arQuarterJSON = `{
  "name": "ar-quarter",
  "description": "quarter-wave MgF2 on glass at 550 nm",
  "design_wavelength_nm": 550,
  "incident": {"index": 1.0},
  "layers": [{"index": 1.38, "thickness_nm": 99.64}],
  "substrate": {"index": 1.5},
  "wavelength_nm": 550,
  "angle_deg": 0,
  "polarization": "average"
}`

// stackBody is a clean /api/stack request: the same optics as the packaged
// example but without the example-only metadata fields.
const stackBody = `{
  "incident": {"index": 1.0},
  "layers": [{"index": 1.38, "thickness_nm": 99.64}],
  "substrate": {"index": 1.5},
  "wavelength_nm": 550,
  "angle_deg": 0,
  "polarization": "average"
}`

const absorberJSON = `{
  "name": "absorber-stack",
  "incident": {"index": 1.0},
  "layers": [{"index": 3.17, "extinction": 3.33, "thickness_nm": 20.0}],
  "substrate": {"index": 1.5},
  "wavelength_nm": 550,
  "angle_deg": 0,
  "polarization": "average"
}`

// postJSON posts a raw body to the test server and returns status + body.
func postJSON(t *testing.T, ts *httptest.Server, path, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(ts.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(data)
}

// TestStackEndpoint checks POST /api/stack end to end: a valid quarter-wave
// request returns the expected reflectance with an energy budget of one.
func TestStackEndpoint(t *testing.T) {
	ts := testServer(t)
	status, body := postJSON(t, ts, "/api/stack", stackBody)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	var res struct {
		Reflection float64 `json:"reflection"`
		EnergySum  float64 `json:"energy_sum"`
		BareRefl   float64 `json:"bare_reflection"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatalf("response not JSON: %v; body: %s", err, body)
	}
	if want := 0.0141; res.Reflection > want+1e-4 || res.Reflection < want-1e-4 {
		t.Errorf("reflection = %v, want ≈ %v", res.Reflection, want)
	}
	if d := res.EnergySum - 1; d > 1e-6 || d < -1e-6 {
		t.Errorf("energy_sum = %v, want 1", res.EnergySum)
	}
	if res.BareRefl <= res.Reflection {
		t.Errorf("bare_reflection = %v should exceed coated reflection %v", res.BareRefl, res.Reflection)
	}
}

// TestSpectrumEndpoint checks POST /api/spectrum returns the requested point
// series with a reflectance minimum near the quarter-wave design.
func TestSpectrumEndpoint(t *testing.T) {
	ts := testServer(t)
	body := `{
	  "incident": {"index": 1.0},
	  "layers": [{"index": 1.38, "thickness_nm": 99.64}],
	  "substrate": {"index": 1.5},
	  "wavelength_min_nm": 400,
	  "wavelength_max_nm": 700,
	  "points": 61,
	  "angle_deg": 0,
	  "polarization": "s"
	}`
	status, raw := postJSON(t, ts, "/api/spectrum", body)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", status, raw)
	}
	var res struct {
		Points []struct {
			Reflection float64 `json:"reflection"`
		} `json:"points"`
		MinReflection struct {
			WavelengthNm float64 `json:"wavelength_nm"`
		} `json:"min_reflection"`
	}
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		t.Fatalf("response not JSON: %v; body: %s", err, raw)
	}
	if len(res.Points) != 61 {
		t.Errorf("point count = %d, want 61", len(res.Points))
	}
	if wl := res.MinReflection.WavelengthNm; wl < 500 || wl > 600 {
		t.Errorf("min reflection at %v nm, want near the 550 nm design", wl)
	}
}

// TestStackEndpointRejectsInvalid checks that a physically impossible request
// returns a 400 error body produced by the backend validation.
func TestStackEndpointRejectsInvalid(t *testing.T) {
	ts := testServer(t)
	bad := strings.Replace(stackBody, `"thickness_nm": 99.64`, `"thickness_nm": -20`, 1)
	status, body := postJSON(t, ts, "/api/stack", bad)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", status, body)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &errBody); err != nil {
		t.Fatalf("error body not JSON: %v; body: %s", err, body)
	}
	if errBody.Error == "" {
		t.Error("error message is empty")
	}
	if !strings.Contains(errBody.Error, "层厚") {
		t.Errorf("error %q does not name the offending field", errBody.Error)
	}
}

// TestUnknownFieldRejected checks that a misspelled JSON key is an explicit
// 400 rather than silently ignored.
func TestUnknownFieldRejected(t *testing.T) {
	ts := testServer(t)
	typo := strings.Replace(stackBody, `"wavelength_nm"`, `"wavelegnth_nm"`, 1)
	status, body := postJSON(t, ts, "/api/stack", typo)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", status, body)
	}
	if !strings.Contains(body, "error") {
		t.Errorf("body has no error field: %s", body)
	}
}

// TestExamplesEndpoint checks the example listing and single-example fetch.
func TestExamplesEndpoint(t *testing.T) {
	ts := testServer(t)
	resp, err := http.Get(ts.URL + "/api/examples")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var list struct {
		Examples []struct {
			Name string `json:"name"`
		} `json:"examples"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("list response not JSON: %v", err)
	}
	found := false
	for _, e := range list.Examples {
		if e.Name == "ar-quarter" {
			found = true
		}
	}
	if !found {
		t.Errorf("example list %+v does not contain ar-quarter", list.Examples)
	}

	resp2, err := http.Get(ts.URL + "/api/examples/ar-quarter")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	data, _ := io.ReadAll(resp2.Body)
	if !strings.Contains(string(data), `"thickness_nm"`) {
		t.Errorf("example body missing thickness: %s", data)
	}

	// Unknown example → structured 404.
	resp3, err := http.Get(ts.URL + "/api/examples/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Errorf("unknown example status = %d, want 404", resp3.StatusCode)
	}
}

// TestWebIndexServed checks that GET / serves the embedded console HTML.
func TestWebIndexServed(t *testing.T) {
	ts := testServer(t)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(data), "fresnel-tm") {
		t.Errorf("index body missing title: %s", data)
	}

	resp2, err := http.Get(ts.URL + "/example/ar-quarter.json")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("example static status = %d, want 200", resp2.StatusCode)
	}
}
