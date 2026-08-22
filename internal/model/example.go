package model

import (
	"encoding/json"
	"fmt"
	"os"
)

// Example is a packaged example problem. Its JSON layout is exactly the body
// of a /api/stack request (embedded StackRequest) plus metadata, so every
// example can be posted straight to the solver without transformation.
//
// DesignWavelengthNm labels the wavelength the coating was designed for
// (e.g. the λ0 of a quarter-wave stack). It is metadata for the frontend and
// the CLI; the physics uses only the embedded request fields.
type Example struct {
	// Name is the file stem, e.g. "ar-quarter".
	Name string `json:"name"`
	// Description explains what the example demonstrates.
	Description string `json:"description,omitempty"`
	// DesignWavelengthNm is the design wavelength in nanometres, if any.
	DesignWavelengthNm float64 `json:"design_wavelength_nm,omitempty"`
	StackRequest
}

// Validate checks the metadata and the embedded request.
func (e Example) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("示例缺少 name 字段")
	}
	return e.StackRequest.Validate()
}

// ParseExample decodes a raw example file (JSON) and validates it.
func ParseExample(data []byte) (*Example, error) {
	var ex Example
	if err := json.Unmarshal(data, &ex); err != nil {
		return nil, fmt.Errorf("示例 JSON 无法解析：%w", err)
	}
	if err := ex.Validate(); err != nil {
		return nil, err
	}
	return &ex, nil
}

// LoadExampleFile reads and parses an example from disk. It is used by the
// -solve CLI mode and by tests that exercise the packaged files.
func LoadExampleFile(path string) (*Example, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取示例 %s 失败：%w", path, err)
	}
	return ParseExample(data)
}

// WithWavelength returns a copy of the example solved at the given wavelength,
// keeping the design wavelength metadata untouched.
func (e Example) WithWavelength(nm float64) *Example {
	cp := e
	cp.StackRequest.WavelengthNm = nm
	return &cp
}
