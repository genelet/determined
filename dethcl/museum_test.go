package dethcl

import (
	"os"
	"testing"

	"github.com/genelet/determined/utils"
	//"github.com/genelet/determined/utils"
)

type museum struct {
	Location string              `hcl:"location"`
	Arts     map[string]*picture `hcl:"arts,block"`
}

func TestMapSliceNew(t *testing.T) {
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

	bs, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("museum.hcl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Write(bs)
}

func TestMapSliceRead(t *testing.T) {
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
	for k, v := range m.Arts {
		for _, d := range v.Drawings {
			t.Logf("%s, %s, %#v", k, v.Name, d)
		}
	}
	t.Errorf("%#v", m)
}
