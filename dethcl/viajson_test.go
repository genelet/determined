package dethcl

import (
	"encoding/json"
	"reflect"
	"testing"
)

type Third struct {
	Cafe   string `hcl:"cafe" json:"cafe"`
	Street string `hcl:"street" json:"street"`
	Number int    `hcl:"number" json:"number"`
}

type Second struct {
	Book   string  `hcl:"book" json:"book"`
	Author string  `hcl:"author" json:"author"`
	Price  float64 `hcl:"price" json:"price"`
	Year   int     `hcl:"year" json:"year"`
}

type First struct {
	Seconds    map[string]*Second `hcl:"seconds,block" json:"seconds"`
	Thirds     []*Third           `hcl:"thirds,block" json:"thirds"`
	Department string             `hcl:"department" json:"department"`
	Score      float64            `hcl:"score" json:"score"`
}

func getFirst() *First {
	third1 := &Third{
		Cafe:   "capccino",
		Street: "hedrid",
		Number: 1,
	}
	third2 := &Third{
		Cafe:   "latte",
		Street: "court",
		Number: 2,
	}
	second1 := &Second{
		Book:   "The Go Programming Language",
		Author: "Alan A. A. Donovan",
		Price:  50.0,
		Year:   2015,
	}
	second2 := &Second{
		Book:   "The essential Go",
		Author: "Bob",
		Price:  20.0,
		Year:   2018,
	}
	return &First{
		Seconds: map[string]*Second{
			"aaa": second1,
			"bbb": second2,
		},
		Thirds: []*Third{
			third1,
			third2,
		},
		Department: "IT",
		Score:      100.0,
	}
}

func TestViaJSONEasy(t *testing.T) {
	first := getFirst()

	bs, err := Marshal(first)
	if err != nil {
		t.Errorf("Marshal error: %v", err)
	}

	first1 := &First{}
	err = Unmarshal(bs, first1)
	if err != nil {
		t.Errorf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(first, first1) {
		t.Errorf("Unmarshal error: %#v => %#v", first, first1)
	}
}

// TestViaJSONHash is a test that hcl-marshals a struct, and hcl-unmarshals the struct to hash,
// then json-marshals the hash to bytes string, ...transfering ..., and json-unmarshal to the struct
// from the json string. The final struct should be equal to the original struct.
func TestViaJSONHash(t *testing.T) {
	first := getFirst()

	bs, err := Marshal(first)
	if err != nil {
		t.Errorf("Marshal error: %v", err)
	}

	hash := make(map[string]any)
	err = Unmarshal(bs, &hash)
	if err != nil {
		t.Errorf("Unmarshal error: %v", err)
	}

	bs1, err := json.Marshal(hash)
	if err != nil {
		t.Errorf("Marshal error: %v", err)
	}

	first1 := &First{}
	err = json.Unmarshal(bs1, first1)
	if err != nil {
		t.Errorf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(first, first1) {
		t.Errorf("Unmarshal error: %#v => %#v", first, first1)
	}
}
