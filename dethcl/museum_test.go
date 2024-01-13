package dethcl

import (
	"os"
	"reflect"
	"testing"

	"github.com/genelet/determined/utils"
	//"github.com/genelet/determined/utils"
)

type museum struct {
	Location string              `hcl:"location"`
	Arts     map[string]*picture `hcl:"arts,block"`
}

func getMuseum() *museum {
	p1 := &picture{
		Name: "peter drawings first",
		Drawings: []inter{
			&square{SX: 1, SY: 2},
			&square{SX: 3, SY: 4},
		},
	}
	p2 := &picture{
		Name: "peter drawings second",
		Drawings: []inter{
			&square{SX: 5, SY: 6},
			&square{SX: 7, SY: 8},
			&square{SX: 9, SY: 10},
			&square{SX: 11, SY: 12},
		},
	}
	p3 := &picture{
		Name: "peter drawings third",
		Drawings: []inter{
			&circle{Radius: 5},
			&circle{Radius: 6},
			&circle{Radius: 7},
		},
	}
	p4 := &picture{
		Name: "peter drawings forth",
		Drawings: []inter{
			&circle{Radius: 8},
			&circle{Radius: 9},
			&circle{Radius: 10},
			&circle{Radius: 11},
			&circle{Radius: 12},
		},
	}
	p5 := &picture{
		Name: "peter drawings fifth",
		Drawings: []inter{
			&square{SX: 13, SY: 14},
			&circle{Radius: 15},
			&square{SX: 16, SY: 17},
		},
	}

	m := &museum{
		Location: "Chicago",
		Arts: map[string]*picture{
			"first":  p1,
			"second": p2,
			"third":  p3,
			"forth":  p4,
			"fifth":  p5,
		},
	}

	return m
}

func TestMapSliceNew(t *testing.T) {
	mRaw := getMuseum()
	bsRaw, err := Marshal(mRaw)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("museum.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Write(bsRaw)

	bs, err := os.ReadFile("museum.hcl")
	if err != nil {
		t.Fatal(err)
	}

	m := &museum{}
	ref := map[string]interface{}{
		"square":  new(square),
		"circle":  new(circle),
		"picture": new(picture),
		"museum":  new(museum),
	}
	spec, err := utils.NewStruct(
		"museum", map[string]interface{}{
			"Arts": map[string][2]interface{}{
				"first": {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square"}},
				},
				"second": {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square", "square", "square"}},
				},
				"third": {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle"}},
				},
				"forth": {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle", "circle", "circle"}},
				},
				"fifth": {"picture", map[string]interface{}{
					"Drawings": []string{"square", "circle", "square"}},
				},
			},
		},
	)
	err = UnmarshalSpec(bs, m, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, getMuseum()) {
		for k, v := range m.Arts {
			for _, d := range v.Drawings {
				t.Logf("%s, %s, %#v", k, v.Name, d)
			}
		}
		t.Errorf("%#v", m)
	}

	os.Remove("museum.hcl")
}

func TestMap2Hcl(t *testing.T) {
	type datacenter struct {
		Alias  string `json:"alias" hcl:"alias,label"`
		Region string `json:"region" hcl:"region"`
	}

	type provider struct {
		Provider map[[2]string]*datacenter `json:"provider" hcl:"provider,block"`
	}

	data2 := `
		   	# default configuration
		   	provider "google" "default" {
		   	  region = "us-central1"
		   	}

		   	# alternate configuration, whose alias is "europe"
		   	provider "google" {
			  alias = "europe"				
		   	  region = "europe-west1"
		   	}

			provider "amazon" {
			  alias = "us"
			  region = "lax-1"
			}

`
	p := new(provider)
	err := UnmarshalSpec([]byte(data2), p, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range p.Provider {
		switch k[0] {
		case "google":
			switch k[1] {
			case "europe":
				if v.Region != "europe-west1" {
					t.Errorf("%v => %#v", k, v)
				}
			default:
				if v.Region != "us-central1" {
					t.Errorf("%v => %#v", k, v)
				}
			}
		case "amazon":
			if !(k[1] == "us" && v.Region == "lax-1") {
				t.Errorf("%v => %#v", k, v)
			}
		default:
		}
	}
}

type museum2 struct {
	Location string                 `hcl:"location"`
	Arts     map[[2]string]*picture `hcl:"arts,block"`
}

func getMuseum2() (*museum2, *museum2) {
	p1 := &picture{
		Name: "peter drawings first",
		Drawings: []inter{
			&square{SX: 1, SY: 2},
			&square{SX: 3, SY: 4},
		},
	}
	p2 := &picture{
		Name: "peter drawings second",
		Drawings: []inter{
			&square{SX: 5, SY: 6},
			&square{SX: 7, SY: 8},
			&square{SX: 9, SY: 10},
			&square{SX: 11, SY: 12},
		},
	}
	p3 := &picture{
		Name: "peter drawings third",
		Drawings: []inter{
			&circle{Radius: 5},
			&circle{Radius: 6},
			&circle{Radius: 7},
		},
	}
	p4 := &picture{
		Name: "peter drawings forth",
		Drawings: []inter{
			&circle{Radius: 8},
			&circle{Radius: 9},
			&circle{Radius: 10},
			&circle{Radius: 11},
			&circle{Radius: 12},
		},
	}
	p5 := &picture{
		Name: "peter drawings fifth",
		Drawings: []inter{
			&square{SX: 13, SY: 14},
			&circle{Radius: 15},
			&square{SX: 16, SY: 17},
		},
	}

	m2 := &museum2{
		Location: "Chicago",
		Arts: map[[2]string]*picture{
			{"first", "x1"}:  p1,
			{"second", "x2"}: p2,
			{"third", "x3"}:  p3,
			{"forth", "x4"}:  p4,
			{"fifth", "x5"}:  p5,
		},
	}

	m20 := &museum2{
		Location: "Chicago",
		Arts: map[[2]string]*picture{
			{"first"}:  p1,
			{"second"}: p2,
			{"third"}:  p3,
			{"forth"}:  p4,
			{"fifth"}:  p5,
		},
	}

	return m2, m20
}

func TestMap2New(t *testing.T) {
	m, m0 := getMuseum2()

	bs, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("museum2.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Write(bs)

	bs0, err := Marshal(m0)
	if err != nil {
		t.Fatal(err)
	}

	f0, err := os.Create("museum20.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f0.Write(bs0)
}

func TestMap2Read(t *testing.T) {
	bs, err := os.ReadFile("museum2.hcl")
	if err != nil {
		t.Fatal(err)
	}

	m2 := &museum2{}
	ref := map[string]interface{}{
		"square":  new(square),
		"circle":  new(circle),
		"picture": new(picture),
		"museum2": new(museum2),
	}
	spec, err := utils.NewStruct(
		"museum2", map[string]interface{}{
			"Arts": map[[2]string][2]interface{}{
				{"first", "x1"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square"}},
				},
				{"second", "x2"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square", "square", "square"}},
				},
				{"third", "x3"}: {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle"}},
				},
				{"forth", "x4"}: {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle", "circle", "circle"}},
				},
				{"fifth", "x5"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "circle", "square"}},
				},
			},
		},
	)
	err = UnmarshalSpec(bs, m2, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	check, _ := getMuseum2()
	if !reflect.DeepEqual(m2, check) {
		for k, v := range m2.Arts {
			for _, d := range v.Drawings {
				t.Logf("%s, %s, %#v", k, v.Name, d)
			}
		}
		t.Errorf("%#v", m2)
	}
}
func TestMap20New(t *testing.T) {
	m, m0 := getMuseum2()

	bs, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("museum2.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Write(bs)

	bs0, err := Marshal(m0)
	if err != nil {
		t.Fatal(err)
	}

	f0, err := os.Create("museum20.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f0.Write(bs0)
}

func TestMap20Read(t *testing.T) {
	bs, err := os.ReadFile("museum20.hcl")
	if err != nil {
		t.Fatal(err)
	}

	m2 := &museum2{}
	ref := map[string]interface{}{
		"square":  new(square),
		"circle":  new(circle),
		"picture": new(picture),
		"museum2": new(museum2),
	}
	spec, err := utils.NewStruct(
		"museum2", map[string]interface{}{
			"Arts": map[[2]string][2]interface{}{
				{"first"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square"}},
				},
				{"second"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "square", "square", "square"}},
				},
				{"third"}: {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle"}},
				},
				{"forth"}: {"picture", map[string]interface{}{
					"Drawings": []string{"circle", "circle", "circle", "circle", "circle"}},
				},
				{"fifth"}: {"picture", map[string]interface{}{
					"Drawings": []string{"square", "circle", "square"}},
				},
			},
		},
	)
	err = UnmarshalSpec(bs, m2, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	_, check := getMuseum2()
	if !reflect.DeepEqual(m2, check) {
		for k, v := range m2.Arts {
			for _, d := range v.Drawings {
				t.Logf("%s, %s, %#v", k, v.Name, d)
			}
		}
		t.Errorf("%#v", m2)
	}

	os.Remove("museum2.hcl")
	os.Remove("museum20.hcl")
}
