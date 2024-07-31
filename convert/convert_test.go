package convert

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/genelet/determined/dethcl"
	"gopkg.in/yaml.v3"
)

func TestYaml2Json(t *testing.T) {
	yaml2(t, "x")
	yaml2(t, "y")
	yaml2(t, "z")
}

func yaml2(t *testing.T, fn string) {
	raw, err := os.ReadFile(fn + ".yaml")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	jsn, err := YAMLToJSON(raw)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	rawjson, err := os.ReadFile(fn + ".json")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if strings.TrimSpace(string(jsn)) != strings.TrimSpace(string(rawjson)) {
		t.Errorf("jsn: %s\n", jsn)
		t.Errorf("raw: %s\n", rawjson)
	}

	hcl, err := os.ReadFile(fn + ".hcl")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	expected, err := YAMLToHCL(raw)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	hclmap := map[string]interface{}{}
	expectedmap := map[string]interface{}{}
	err = dethcl.Unmarshal(hcl, &hclmap)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	err = dethcl.Unmarshal(expected, &expectedmap)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if !reflect.DeepEqual(hclmap, expectedmap) {
		t.Errorf("hcl: %#v\n", hclmap)
		t.Errorf("expected: %#v\n", expectedmap)
	}
}

func TestHcl2Json(t *testing.T) {
	hcl2(t, "x")
	hcl2(t, "y")
	hcl2(t, "z")
}

func hcl2(t *testing.T, fn string) {
	raw, err := os.ReadFile(fn + ".hcl")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	jsn, err := HCLToJSON(raw)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	rawjson, err := os.ReadFile(fn + ".json")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if strings.TrimSpace(string(jsn)) != strings.TrimSpace(string(rawjson)) {
		t.Errorf("jsn: %s\n", jsn)
		t.Errorf("raw: %s\n", rawjson)
	}

	rawyml, err := os.ReadFile(fn + ".yaml")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	expected, err := HCLToYAML(raw)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	ymlmap := map[string]interface{}{}
	expectedmap := map[string]interface{}{}
	err = yaml.Unmarshal(rawyml, &ymlmap)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	err = yaml.Unmarshal(expected, &expectedmap)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if !reflect.DeepEqual(ymlmap, expectedmap) {
		t.Errorf("yaml: %#v\n", ymlmap)
		t.Errorf("expected: %#v\n", expectedmap)
	}
}

func TestHcl2self(t *testing.T) {
	hcl2self(t, "x")
	hcl2self(t, "y")
	hcl2self(t, "z")
}

func hcl2self(t *testing.T, fn string) {
	raw, err := os.ReadFile(fn + ".hcl")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	hash := map[string]interface{}{}
	err = dethcl.Unmarshal(raw, &hash)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	bs, err := dethcl.Marshal(hash)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}

	hash1 := map[string]interface{}{}
	err = dethcl.Unmarshal(bs, &hash1)
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	if !reflect.DeepEqual(hash, hash1) {
		t.Errorf("hash: %#v\n", hash)
		t.Errorf("hash1: %#v\n", hash1)
	}
}
