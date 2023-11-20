package utils

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/genelet/determined/dethcl"
)

func TestYaml2Json(t *testing.T) {
	yaml2(t, "x")
	yaml2(t, "y")
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
	err = dethcl.Unmarshal(refresh(hcl), &hclmap)
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
