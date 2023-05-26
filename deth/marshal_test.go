package deth

import (
	"testing"
)

func TestMHclSimple(t *testing.T) {
	data1 := `
	radius = 1.0
`
	c := new(circle)
	err := HclUnmarshal([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := HclMarshal(c)
	if err != nil { t.Fatal(err) }
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
	err := HclUnmarshal([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := HclMarshal(c)
	if err != nil { t.Fatal(err) }
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
	endpoint, err := NewStruct(
		"geo", map[string]interface{}{"Shape": "circle"})
	err = HclUnmarshal([]byte(data1), g, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := HclMarshal(g)
	if err != nil { t.Fatal(err) }
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
	endpoint, err = NewStruct(
		"geo", map[string]interface{}{"Shape": "square"})
	err = HclUnmarshal([]byte(data2), g, endpoint, ref)
	if err != nil { t.Fatal(err) }
	bs, err = HclMarshal(g)
	if err != nil { t.Fatal(err) }
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
	bs, err = HclMarshal(p)
	if err != nil { t.Fatal(err) }
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
	endpoint, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"moresquare", "moresquare"}})
	if err != nil {
		t.Fatal(err)
	}
	ref["moresquare"] = &moresquare{}
	err = HclUnmarshal([]byte(data5), p, endpoint, ref)
	if err != nil { t.Fatal(err) }

	bs, err = HclMarshal(p)
	if err != nil { t.Fatal(err) }
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
	endpoint, err := NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": []string{"square", "square"}})
			//"Shapes": []string{"square", "square"}})
	ref := map[string]interface{}{"square": new(square)}
	err = HclUnmarshal([]byte(data3), g, endpoint, ref)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := HclMarshal(g)
	if err != nil { t.Fatal(err) }
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

	bs, err := HclMarshal(c)
	if err != nil { t.Fatal(err) }
	if string(bs) != `
brand {

  geo {
    name = "peter shape"

    shape {
      radius = 1
    }
  }

  toy_name = "roblox"
  price    = 99.9000015258789
}

age = 5
` {
		t.Errorf("'%s'", bs)
    }
}
