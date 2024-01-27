package dethcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Returns nil and nil error when input is nil
func TestMarshal_NilInput(t *testing.T) {
	result, err := Marshal(nil)
	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestMarshal_StructInput(t *testing.T) {
	type A struct {
		X string
	}
	type B struct {
		A
		Z A
		Y int
	}
	input := B{A: A{X: "peter"}, Z: A{X: "Marcus"}, Y: 2}
	result, err := Marshal(input)
	expected := []byte(`  x = "peter"
  y = 2
  z {
    x = "Marcus"
  }`)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}

// Returns expected byte slice and nil error when input is valid
func TestMarshal_ValidInput(t *testing.T) {
	input := struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  30,
	}
	expected := []byte(`  name = "John"
  age  = 30`)
	result, err := Marshal(input)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}

// Returns error when input is invalid
func TestMarshal_InvalidInput(t *testing.T) {
	input := make(chan int)
	result, err := Marshal(input)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}

// Handles recursive data structures without infinite loop
func TestMarshal_RecursiveDataStructure(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}
	node1 := &Node{Value: 1}
	node2 := &Node{Value: 2}
	node1.Next = node2

	result, err := Marshal(node1)
	expected := []byte(`  value = 1
  next {
    value = 2
  }`)
	assert.NotNil(t, result)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}

// Handles circular references without infinite loop
func TestMarshal_CircularReferences(t *testing.T) {
	type Person struct {
		Name    string
		Friends []*Person
	}
	john := &Person{Name: "John"}
	jane := &Person{Name: "Jane"}
	sam := &Person{Name: "Sam"}
	jack := &Person{Name: "Jack"}
	jack.Friends = append(jack.Friends, sam)
	john.Friends = append(john.Friends, jane)
	john.Friends = append(john.Friends, jack)

	result, err := Marshal(john)
	assert.NotNil(t, result)
	expected := []byte(`  name = "John"
  friends {
    name = "Jane"
  }
  friends {
    name = "Jack"
    friends {
      name = "Sam"
    }
  }`)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}

// Handles unexported fields in structs
func TestMarshal_UnexportedFields(t *testing.T) {
	type Person struct {
		name string
		Age  int
	}
	john := Person{name: "John", Age: 30}

	result, err := Marshal(john)
	assert.NotNil(t, result)
	assert.Nil(t, err)
}

// Handles embedded structs and pointers to structs
func TestMarshal_EmbeddedStructsAndPointers(t *testing.T) {
	type InnerStruct struct {
		Field1 string
		Field2 int
		Field3 interface{}
		Field4 bool
		Field5 *string
	}

	type OuterStruct struct {
		Inner *InnerStruct
	}

	input := &OuterStruct{
		Inner: &InnerStruct{
			Field1: "value1",
			Field2: 123,
		},
	}

	expected := []byte(`  inner {
    field1 = "value1"
    field2 = 123
  }`)

	result, err := Marshal(input)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}

// Handles maps and slices of basic types
func TestMarshal_MapsAndSlicesOfBasicTypes(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	expected := []byte(`
  key1 = "value1"
  key2 = 123
  key3 = true
`)

	result, err := Marshal(input)
	assert.Equal(t, string(expected), string(result))
	assert.Nil(t, err)
}
