package dethcl

import (
	"testing"
)

func TestHclSimple(t *testing.T) {
	data1 := `
	radius = 1.234
`
	c := new(circle)
	err := Unmarshal([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}

func TestHclShape1(t *testing.T) {
	data1 := `
	name = "peter shape"
	shape {
		radius = 1.234
	}
`
	g := &geo{}
	c := &circle{}
	ref := map[string]interface{}{"circle": c}
	spec, err := NewStruct(
		"geo", map[string]interface{}{"Shape": "circle"})
	err = Unmarshal([]byte(data1), g, spec, ref)
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
	spec, err = NewStruct(
		"geo", map[string]interface{}{"Shape": "square"})
	err = Unmarshal([]byte(data2), g, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "peter shape" || g.Shape.(*square).SX != 5 {
		t.Errorf("%#v", g)
	}
}

func TestHclShape2(t *testing.T) {
	data4 := `
	name = "peter drawings"
	drawings {
		sx=55
		sy=66
	}
	drawings {
		sx=7
		sy=8
	}
`
	p := &picture{}
	spec, err := NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		t.Fatal(err)
	}

	g := &geo{}
	s := &square{}
	c := &circle{}
	ref := map[string]interface{}{"geo": g, "circle": c, "square": s}
	err = Unmarshal([]byte(data4), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings := p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*square).SX != 55 ||
		drawings[1].(*square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*square))
		t.Errorf("%#v", drawings[1].(*square))
	}

	data5 := `
    name = "peter drawings"
    drawings "abc1" def1 {
        sx=5
        sy=6
    }
    drawings abc2 "def2" {
        sx=7
        sy=8
    }
`
	p = &picture{}
	spec, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"moresquare", "moresquare"}})
	if err != nil {
		t.Fatal(err)
	}
	ref["moresquare"] = &moresquare{}
	err = Unmarshal([]byte(data5), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings = p.Drawings
	if p.Name != "peter drawings" ||
		drawings[0].(*moresquare).Morename1 != "abc1" ||
		drawings[0].(*moresquare).SX != 5 ||
		drawings[1].(*moresquare).Morename2 != "def2" ||
		drawings[1].(*moresquare).SX != 7 {
		t.Errorf("%#v", drawings[0].(*moresquare))
		t.Errorf("%#v", drawings[1].(*moresquare))
	}

	bs, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `name = "peter drawings"
drawings abc1 def1 {
  sx = 5
  sy = 6
}

drawings abc2 def2 {
  sx = 7
  sy = 8
}

` {
		t.Errorf("%s", bs)
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
	spec, err := NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": []string{"square", "square"}})
	//"Shapes": []string{"square", "square"}})
	ref := map[string]interface{}{"square": new(square)}
	err = Unmarshal([]byte(data3), g, spec, ref)
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
	spec, err := NewStruct(
		"child1", map[string]interface{}{
			"Brand": [2]interface{}{
				"toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"geo", map[string]interface{}{"Shape": "circle"}}}}})
	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "toy": &toy{}}

	c := new(child)
	err = Unmarshal([]byte(data1), c, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if c.Age != 5 || c.Brand.Geo.Shape.(*circle).Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}
