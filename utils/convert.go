package utils

import (
	"encoding/json"
	"strings"

	"github.com/genelet/determined/dethcl"
	"gopkg.in/yaml.v3"
)

// JSONToYAML converts JSON to YAML.
func JSONToYAML(raw []byte) ([]byte, error) {
	obj := map[string]interface{}{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return yaml.Marshal(obj)
}

// YAMLToJSON converts YAML to JSON.
func YAMLToJSON(raw []byte) ([]byte, error) {
	obj := map[string]interface{}{}
	err := yaml.Unmarshal(raw, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

// JSONToHCL converts JSON to HCL.
func JSONToHCL(raw []byte) ([]byte, error) {
	obj := map[string]interface{}{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return dethcl.Marshal(obj)
}

// refresh refreshes the raw string to a valid HCL string.
func refresh(raw []byte) []byte {
	str := strings.TrimSpace(string(raw))
	if str[0:1] != `{` {
		str = "{\n" + str + "\n}"
	}
	return []byte(str)
}

// HCLToJSON converts HCL to JSON.
func HCLToJSON(raw []byte) ([]byte, error) {
	raw = refresh(raw)
	obj := map[string]interface{}{}
	if err := dethcl.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

// YAMLToHCL converts YAML to HCL.
func YAMLToHCL(raw []byte) ([]byte, error) {
	obj := map[string]interface{}{}
	err := yaml.Unmarshal(raw, &obj)
	if err != nil {
		return nil, err
	}
	return dethcl.Marshal(obj)
}

// HCLToYAML converts HCL to YAML.
func HCLToYAML(raw []byte) ([]byte, error) {
	raw = refresh(raw)
	obj := map[string]interface{}{}
	if err := dethcl.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return yaml.Marshal(obj)
}
