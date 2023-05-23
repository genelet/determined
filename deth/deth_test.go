package deth

import (
	"testing"
)

func TestHclSimple(t *testing.T) {
	data1 := `
	radius = 1.234
`
	c := new(circle)
	err := HclUnmarshal([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}

func TestHclShape(t *testing.T) {
	data1 := `
	name = "peter shape"
	shape {
		radius = 1.234
	}
`
	g := &geo{}
	c := &circle{}
	ref := map[string]interface{}{"circle": c}
	endpoint, err := NewStruct(
		"geo", map[string]interface{}{"Shape": "circle"})
	err = HclUnmarshal([]byte(data1), g, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "peter shape" || g.Shape.(*circle).Radius != 1.234 {
		t.Errorf("%#v", g)
	}

	data2 := `
	name = "peter shape"
	shape {
    	sx = 5
    	sy = 6
	}
`
	g = &geo{}
	s := &square{}
	ref = map[string]interface{}{"circle": c, "square": s}
	endpoint, err = NewStruct(
		"geo", map[string]interface{}{"Shape": "square"})
	err = HclUnmarshal([]byte(data2), g, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "peter shape" || g.Shape.(*square).SX != 5 {
		t.Errorf("%#v", g)
	}

	data4 := `
	name = "peter drawings"
	drawings {
		sx=5
		sy=6
	}
	drawings {
		sx=7
		sy=8
	}
`
	p := &picture{}
	endpoint, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = HclUnmarshal([]byte(data4), p, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings := p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*square).SX != 5 ||
		drawings[1].(*square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*square))
		t.Errorf("%#v", drawings[1].(*square))
	}
}

func TestHash(t *testing.T) {
	data3 := `
	name = "peter shapes"
	shapes obj5 {
		sx = 5
		sy = 6
	}
	shapes obj7 {
		sx = 7
		sy = 8
	}
`
	g := &geometry{}
	endpoint, err := NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": [][1]string{{"square"}, {"square"}}})
			//"Shapes": []string{"square", "square"}})
	ref := map[string]interface{}{"square": new(square)}
	err = HclUnmarshal([]byte(data3), g, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	shapes := g.Shapes
	if g.Name != "peter shapes" ||
		shapes["obj5"].(*square).SX != 5 ||
		shapes["obj7"].(*square).SX != 7 {
		t.Errorf("%#v", shapes["obj5"].(*square))
		t.Errorf("%#v", shapes["obj7"].(*square))
	}
}

func TestHclChild(t *testing.T) {
	data1 := `
age = 5
brand {
	toy_name = "roblox"
	price = 99.9
	geo {
		name = "peter shape"
		shape {
    		radius = 1.234
		}
	}
}
`
	endpoint, err := NewStruct(
		"child1", map[string]interface{}{
			"Brand": [2]interface{}{
				"toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"geo", map[string]interface{}{"Shape": "circle"}}}}})
	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "toy": &toy{}}

	c := new(child)
	err = HclUnmarshal([]byte(data1), c, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if c.Age != 5 || c.Brand.Geo.Shape.(*circle).Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}
