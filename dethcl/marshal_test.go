package dethcl

import (
	"testing"
)

func TestMHclSimple(t *testing.T) {
	data1 := `
	radius = 1.0
`
	c := new(circle)
	err := Unmarshal([]byte(data1), c)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "radius = 1\n" {
		t.Errorf("%s", bs)
	}
}

func TestMHclSimpleMore(t *testing.T) {
	data1 := `
	radius = 1.0
arr1 = ["abc", "def"]
arr2 = [123, 4356]
arr3 = [true, false, true]
`
	c := new(circlemore)
	err := Unmarshal([]byte(data1), c)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `radius = 1
arr1   = ["abc", "def"]
arr2   = [123, 4356]
arr3   = [true, false, true]
` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclShape(t *testing.T) {
	data1 := `
	name = "peter shape"
	shape {
		radius = 1.0
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
	bs, err := Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `name = "peter shape"
shape {
	radius = 1
}

` {
		t.Errorf("'%s'", bs)
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
	bs, err = Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `name = "peter shape"
shape {
	sx = 5
	sy = 6
}

` {
		t.Errorf("'%s'", bs)
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
	spec, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = Unmarshal([]byte(data4), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	bs, err = Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `name = "peter drawings"
drawings {
	sx = 5
	sy = 6
}

drawings {
	sx = 7
	sy = 8
}

` {
		t.Errorf("'%s'", bs)
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

	bs, err = Marshal(p)
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
		t.Errorf("'%s'", bs)
	}
}

func TestMHash(t *testing.T) {
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

	bs, err := Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `name = "peter shapes"
shapes obj5 {
	sx = 5
	sy = 6
}

shapes obj7 {
	sx = 7
	sy = 8
}

` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclChild(t *testing.T) {
	data1 := `
age = 5
brand {
	toy_name = "roblox"
	price = 99.9
	geo {
		name = "peter shape"
		shape {
    		radius = 1.0
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

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `age = 5
brand {
	toy_name = "roblox"
	price    = 99.9000015258789
	geo {
		name = "peter shape"
		shape {
			radius = 1
		}
		
	}
	
}

` {
		t.Errorf("'%s'", bs)
	}
}
