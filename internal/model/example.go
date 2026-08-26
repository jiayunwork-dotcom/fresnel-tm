package model

import (
	"encoding/json"
	"fmt"
	"os"
)

type Example struct {
	Name               string  `json:"name"`
	Description        string  `json:"description,omitempty"`
	DesignWavelengthNm float64 `json:"design_wavelength_nm,omitempty"`
	StackRequest
}

func (e Example) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("示例缺少 name 字段")
	}
	return e.StackRequest.Validate()
}

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

func LoadExampleFile(path string) (*Example, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取示例 %s 失败：%w", path, err)
	}
	return ParseExample(data)
}

func (e Example) WithWavelength(nm float64) *Example {
	cp := e
	cp.StackRequest.WavelengthNm = nm
	return &cp
}
