package dethcl

// Unmarshals a valid HCL file into a struct
import (
	"testing"
)

func TestUnmarshal_ValidHCLFile_Struct(t *testing.T) {
	// Prepare test data
	hclData := []byte(`
      x = "peter"
      y = 2
      z {
        x = "Marcus"
      }
      `)
	type A struct {
		X string `hcl:"x"`
	}
	type B struct {
		A
		Z A   `hcl:"z,block"`
		Y int `hcl:"y"`
	}

	var result B

	// Call the function
	err := Unmarshal(hclData, &result)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if result.X != "peter" ||
		result.A.X != "peter" ||
		result.Y != 2 ||
		result.Z.X != "Marcus" {
		t.Errorf("result: %v", result.A)
		t.Errorf("result: %v", result.Z)
		t.Errorf("result: %v", result.Y)
		t.Errorf("result: %v", result.X)
	}
}

// Unmarshals a valid HCL file into a map
func TestUnmarshal_ValidHCLFile_Map(t *testing.T) {
	hclData := []byte(`
     key1 = "value1"
     key2 = 123
     key3 = true
   
       // HCL data here
     `)

	result := map[string]interface{}{}
	err := Unmarshal(hclData, &result)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if result["key1"] != "value1" ||
		result["key2"] != 123 ||
		result["key3"] != true {
		t.Errorf("result: %#v", result)
	}
}

func TestUnmarshal_ValidHCLFileWithLabels_Struct(t *testing.T) {
	// Prepare test data
	hclData := []byte(`
       // HCL data here
       x = "peter"
       y = "marcus"
       z = 2
     `)

	type A struct {
		X string `hcl:"x,label"`
		Y string `hcl:"y,label"`
		Z int    `hcl:"z,optional"`
	}

	var result A
	labels := []string{"label1", "label2"}
	err := Unmarshal(hclData, &result, labels...)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if result.X != "peter" ||
		result.Y != "marcus" ||
		result.Z != 2 {
		t.Errorf("result: %#v", result)
	}
}

func TestUnmarshal_ValidHCLFileWithLabels_Map(t *testing.T) {
	hclData := []byte(`
  c = true
  goal "peter" "marcus" {
    z = 2
  }
`)
	type A struct {
		X string `hcl:"x,label"`
		Y string `hcl:"y,label"`
		Z int    `hcl:"z,optional"`
	}
	type B struct {
		C    bool            `hcl:"c"`
		Goal map[[2]string]A `hcl:"goal,block"`
	}

	var result2 B
	err := Unmarshal(hclData, &result2)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if result2.C != true ||
		result2.Goal[[2]string{"peter", "marcus"}].X != "peter" ||
		result2.Goal[[2]string{"peter", "marcus"}].Y != "marcus" ||
		result2.Goal[[2]string{"peter", "marcus"}].Z != 2 {
		t.Errorf("result: %#v", result2)
	}

	type BB struct {
		C    bool `hcl:"c"`
		Goal []A  `hcl:"goal,block"`
	}
	var result3 BB
	err = Unmarshal(hclData, &result3)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if result3.C != true ||
		result3.Goal[0].X != "peter" ||
		result3.Goal[0].Y != "marcus" ||
		result3.Goal[0].Z != 2 {
		t.Errorf("result: %#v", result3)
	}
}
