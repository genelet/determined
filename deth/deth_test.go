package det

import (
	"testing"
)

func TestHclSimple(t *testing.T) {
	data1 := `
	radius = 1.234
`
	c := new(Circle)
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
	geo := &Geo{}
	c := &Circle{}
	ref := map[string]interface{}{"Circle": c}
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{"Shape": "Circle"})
	err = HclUnmarshal([]byte(data1), geo, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if geo.Name != "peter shape" || geo.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", geo)
	}

	data2 := `
	name = "peter shape"
	shape {
    	sx = 5
    	sy = 6
	}
`
	geo = &Geo{}
	s := &Square{}
	ref = map[string]interface{}{"Circle": c, "Square": s}
	endpoint, err = NewStruct(
		"Geo", map[string]interface{}{"Shape": "Square"})
	err = HclUnmarshal([]byte(data2), geo, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if geo.Name != "peter shape" || geo.Shape.(*Square).SX != 5 {
		t.Errorf("%#v", geo)
	}

	data3 := `
	name =  "peter shapes"
	shapes {
		obj5 = {
			sx = 5
			sy = 6
		}
		obj7 = {
			sx = 7
			sy = 8
		}
	}
`
	geometry := &Geometry{}
	endpoint, err = NewStruct(
		"Geometry", map[string]interface{}{
			"Shapes": map[string]string{
				"obj5": "Square",
				"obj7": "Square"}})
	err = HclUnmarshal([]byte(data3), geometry, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	shapes := geometry.Shapes
	if geometry.Name != "peter shapes" ||
		shapes["obj5"].(*Square).SX != 5 ||
		shapes["obj7"].(*Square).SX != 7 {
		t.Errorf("%#v", shapes["obj5"].(*Square))
		t.Errorf("%#v", shapes["obj7"].(*Square))
	}

	geometry = &Geometry{}
	endpoint, err = NewStruct(
		"Geometry", map[string]interface{}{
			"Shapes": map[string]string{
				"obj7": "Square"}}) // in case of less items, use the first one
	err = HclUnmarshal([]byte(data3), geometry, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	shapes = geometry.Shapes
	if geometry.Name != "peter shapes" ||
		shapes["obj5"].(*Square).SX != 5 ||
		shapes["obj7"].(*Square).SX != 7 {
		t.Errorf("%#v", shapes["obj5"].(*Square))
		t.Errorf("%#v", shapes["obj7"].(*Square))
	}

	data4 := `
	name = "peter drawings"
	drawings = [
		{ sx=5, sy=6 },
		{ sx=7, sy=8 }
	]
`
	picture := &Picture{}
	endpoint, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"Square", "Square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = HclUnmarshal([]byte(data4), picture, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings := picture.Drawings
	if picture.Name != "peter drawings" ||
		drawings[0].(*Square).SX != 5 ||
		drawings[1].(*Square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*Square))
		t.Errorf("%#v", drawings[1].(*Square))
	}

	picture = &Picture{}
	endpoint, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"Square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = HclUnmarshal([]byte(data4), picture, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	drawings = picture.Drawings
	if picture.Name != "peter drawings" ||
		drawings[0].(*Square).SX != 5 ||
		drawings[1].(*Square).SX != 7 {
		t.Errorf("%#v", drawings[0].(*Square))
		t.Errorf("%#v", drawings[1].(*Square))
	}
}

func TestHclToy1(t *testing.T) {
	data1 := `
age = 5
brand = {
	toy_name = "roblox"
	price = 99.9
	geo = {
		name = "peter shape"
		shape = {
    		radius = 1.234
		}
	}
}
`
	endpoint, err := NewStruct(
		"Child", map[string]interface{}{
			"Brand": [2]interface{}{
				"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Circle"}}}}})
	ref := map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

	child := new(Child1)
	err = HclUnmarshal([]byte(data1), child, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	if child.Age != 5 || child.Brand.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", child)
	}
}
