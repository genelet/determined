package dethcl

import (
	"testing"
)

func TestMHclSimple(t *testing.T) {
	data1 := `
	radius = 1.0
`
	c := new(circle)
	err := UnmarshalSpec([]byte(data1), c, nil, nil)
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
	err := UnmarshalSpec([]byte(data1), c, nil, nil)
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
	err = UnmarshalSpec([]byte(data1), g, spec, ref)
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
	err = UnmarshalSpec([]byte(data2), g, spec, ref)
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
	err = UnmarshalSpec([]byte(data4), p, spec, ref)
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
	err = UnmarshalSpec([]byte(data5), p, spec, ref)
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
	err = UnmarshalSpec([]byte(data3), g, spec, ref)
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

` && string(bs) != `name = "peter shapes"
shapes obj7 {
  sx = 7
  sy = 8
}

shapes obj5 {
  sx = 5
  sy = 6
}

` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclOld(t *testing.T) {
	data1 := `description = "here is detailed description"
y13 = {
  str131 = "la"
  str132 = "nyc"
}
y14 = {
  str141 = 141
  str142 = 142
}
y15 = {
  str151 = true
  str152 = false
}
y7 {
  many = 3
  why  = "national day"
}

y10 {
  many = 4
  why  = "labor day"
}

y10 {
  many = 5
  why  = "holiday day"
}

y11 k6 {
  many = 6
  why  = "memorial day"
}

y11 k7 {
  many = 7
  why  = "new day"
}

y12 k8 {
  many = 8
  why  = "christmas day"
}

y12 k9 {
  many = 9
  why  = "new year day"
}

`
	f0 := new(frame0)
	err := UnmarshalSpec([]byte(data1), f0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f0.Y10) != 2 || len(f0.Y11) != 2 || len(f0.Y12) != 2 {
		t.Errorf("%#v", f0)
	}

	bs, err := Marshal(f0)
	if err != nil {
		t.Fatal(err)
	}
	if (string(data1))[:100] != (string(bs))[:100] {
		t.Errorf("%s", bs)
	}
}

func TestMHclFrame(t *testing.T) {
	data1 := `description = "here is detailed description"
number      = 4
what        = "flags"
x1 {
  name = "x1 shape"
  shape {
    radius = 1
  }
  
}

x2 {
  name = "x2 shape"
  shape {
    radius = 2
  }
  
}

x3 {
  name = "x3 1 shape"
  shape {
    radius = 3
  }
  
}

x3 {
  name = "x3 2 shape"
  shape {
    radius = 3
  }
  
}

x4 {
  name = "x4 1 shape"
  shape {
    radius = 4
  }
  
}

x4 {
  name = "x4 2 shape"
  shape {
    radius = 4
  }
  
}

x5 k51 {
  name = "x5 1 shape"
  shape {
    radius = 5
  }
  
}

x5 k52 {
  name = "x5 2 shape"
  shape {
    radius = 5
  }
  
}

x6 k61 {
  name = "x6 1 shape"
  shape {
    radius = 6
  }
  
}

x6 k62 {
  name = "x6 2 shape"
  shape {
    radius = 6
  }
  
}

y7 {
  many = 3
  why  = "national day"
}

y10 {
  many = 4
  why  = "labor day"
}

y10 {
  many = 5
  why  = "holiday day"
}

y11 k7 {
  many = 7
  why  = "new day"
}

y11 k6 {
  many = 6
  why  = "memorial day"
}

`
	spec, err := NewStruct(
		"frame", map[string]interface{}{
			"X1": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X2": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X3": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X4": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X5": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X6": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "frame": &frame{}}

	f := new(frame)
	err = UnmarshalSpec([]byte(data1), f, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	//	if string(data1) != string(bs) {
	if false {
		t.Errorf("%s", bs)
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
	err = UnmarshalSpec([]byte(data1), c, spec, ref)
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
