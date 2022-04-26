package determined

import (
	"encoding/json"
	"testing"
)

type I interface {
	Area() float32
}

type Square struct {
	SX int `json:"sx"`
	SY int `json:"sy"`
}

func (self *Square) Area() float32 {
	return float32(self.SX * self.SY)
}

type Circle struct {
	Radius float32 `json:"radius"`
}

func (self *Circle) Area() float32 {
	return 3.14159 * self.Radius
}

type Cubic struct {
	Size int `json:"size"`
}

func (self *Cubic) Area() float32 {
	return 6.0 * float32(self.Size*self.Size)
}

func TestJsonSimple(t *testing.T) {
	data1 := `{
	"radius":1.234
}`
	c := new(Circle)
	err := JsonUnmarshal([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Radius != 1.234 {
		t.Errorf("%#v", c)
	}
}

type Geo struct {
	Name  string `json:"name"`
	Shape I      `json:"shape"`
}

type Geometry struct {
	Name   string       `json:"name"`
	Shapes map[string]I `json:"shapes"`
}

type Picture struct {
	Name     string `json:"name"`
	Drawings []I    `json:"drawings"`
}

func TestJsonShape(t *testing.T) {
	data1 := `{
	"name": "peter shape",
	"shape": {
    	"radius":1.234
	}
}`
	geo := &Geo{}
	c := &Circle{}
	ref := map[string]interface{}{"Circle":c}
	determined := &Determined{MetaType: METASingle, SingleName: "Circle"}
	dmap := DeterminedMap(map[string]*Determined{"Shape": determined})
	err := JsonUnmarshal([]byte(data1), geo, dmap, ref)
	if err != nil {
		t.Fatal(err)
	}
	if geo.Name != "peter shape" || geo.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", geo)
	}

	data2 := `{
	"name": "peter shape",
	"shape": {
    	"sx":5,
    	"sy":6
	}
}`
	geo = &Geo{}
	s := &Square{}
	ref = map[string]interface{}{"Circle":c, "Square":s}
	determined = &Determined{MetaType: METASingle, SingleName: "Square"}
	dmap = DeterminedMap(map[string]*Determined{"Shape": determined})
	err = JsonUnmarshal([]byte(data2), geo, dmap, ref)
	if err != nil {
		t.Fatal(err)
	}
	if geo.Name != "peter shape" || geo.Shape.(*Square).SX != 5 {
		t.Errorf("%#v", geo)
	}

	data3 := `{
	"name": "peter shapes",
	"shapes": {
		"obj5" : { "sx":5, "sy":6 },
		"obj7" : { "sx":7, "sy":8 }
	}
}`
	geometry := &Geometry{}
	s = &Square{}
	determined = &Determined{MetaType: METAMapSingle, SingleName: "Square"}
	dmap = DeterminedMap(map[string]*Determined{"Shapes": determined})
	err = JsonUnmarshal([]byte(data3), geometry, dmap, ref)
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

	data4 := `{
	"name": "peter drawings",
	"drawings": [
		{ "sx":5, "sy":6 },
		{ "sx":7, "sy":8 }
	]
}`
	picture := &Picture{}
	s = &Square{}
	determined = &Determined{MetaType: METASliceSingle, SingleName: "Square"}
	dmap = DeterminedMap(map[string]*Determined{"Drawings": determined})
	err = JsonUnmarshal([]byte(data4), picture, dmap, ref)
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
	err = json.Unmarshal([]byte(data4), picture)
	if err == nil || err.Error() != "json: cannot unmarshal object into Go struct field Picture.drawings of type determined.I" {
		t.Fatal(err)
	}
}

type Toy struct {
	Geo     `json:"geo"`
	ToyName string  `json:"toy_name"`
	Price   float32 `json:"price"`
}

func (self *Toy) ImportPrice(rate float32) float32 {
	return rate * 0.7 * self.Price
}

type Child struct {
	Toy `json:"toy"`
	Age int `json:"age"`
}

type Adult struct {
	Toys []*Toy `json:"toys"`
	Family bool `json:"family"`
	Lastname string `json:"lastname"`
}

func TestJsonToy(t *testing.T) {
	data0 := `{
    "Toy": {
        "meta_type": 1,
        "single_name": "Toy",
        "single_field": {
            "Geo": {
                "meta_type": 1,
                "single_name": "Geo",
                "single_field": {
                    "Shape": {
                        "meta_type": 1,
                        "single_name": "Circle"
                    }
                }
            }
        }
    }
}`

	data1 := `{
"age":5,
"toy":{
	"toy_name":"roblox",
	"price":99.9,
	"geo":{
		"name": "peter shape",
		"shape": {
    		"radius":1.234
		}
	}
}
}`

	sd   := &Determined{MetaType: METASingle, SingleName: "Circle"}
	smap := DeterminedMap(map[string]*Determined{"Shape": sd})
	gd   := &Determined{MetaType: METASingle, SingleName: "Geo", SingleField: smap}
	gmap := DeterminedMap(map[string]*Determined{"Geo": gd})
	td   := &Determined{MetaType: METASingle, SingleName: "Toy", SingleField: gmap}
	tmap := DeterminedMap(map[string]*Determined{"Toy": td})
	bs, err := json.MarshalIndent(tmap, "", "    ")
	if err != nil { t.Fatal(err) }
	if string(bs) != data0 {
	t.Errorf("%s", bs)
	t.Errorf("%s", bs)
	}

	dmap := DeterminedMap{}
	err = json.Unmarshal([]byte(data0), &dmap)
	if err != nil { t.Fatal(err) }
	ref := map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

	child := new(Child)
	err = JsonUnmarshal([]byte(data1), child, dmap, ref)
	if err != nil { t.Fatal(err) }
	if child.Age != 5 || child.Toy.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", child)
	}

	data1 = `{
"family":true,
"lastname":"Bizhang",
"toys":[{
	"toy_name":"roblox",
	"price":99.9,
	"geo":{
		"name": "peter shape",
		"shape": {
    		"radius":1.234
		}
	}
},{
	"toy_name":"minecraft",
	"price":199.9,
	"geo":{
		"name": "marcus shape",
		"shape": {
    		"sx":134,
			"sy":567
		}
	}
}]
}`

	data0 = `{
  "Toys": {
    "meta_type": 4,
    "slice_name": [
      "Toy",
      "Toy"
    ],
    "slice_field": [
      {
        "Geo": {
          "meta_type": 1,
          "single_name": "Geo",
          "single_field": {
            "Shape": {
              "meta_type": 1,
              "single_name": "Circle"
            }
          }
        }
      },
      {
        "Geo": {
          "meta_type": 1,
          "single_name": "Geo",
          "single_field": {
            "Shape": {
              "meta_type": 1,
              "single_name": "Square"
            }
          }
        }
      }
    ]
  }
}`
	sd1   := &Determined{MetaType: METASingle, SingleName: "Circle"}
	smap1 := DeterminedMap(map[string]*Determined{"Shape": sd1})
	gd1   := &Determined{MetaType: METASingle, SingleName: "Geo", SingleField: smap1}
	gmap1 := DeterminedMap(map[string]*Determined{"Geo": gd1})

	sd2   := &Determined{MetaType: METASingle, SingleName: "Square"}
	smap2 := DeterminedMap(map[string]*Determined{"Shape": sd2})
	gd2   := &Determined{MetaType: METASingle, SingleName: "Geo", SingleField: smap2}
	gmap2 := DeterminedMap(map[string]*Determined{"Geo": gd2})

	td1   := &Determined{MetaType: METASlice, SliceName: []string{"Toy", "Toy"}, SliceField: []DeterminedMap{gmap1, gmap2}}
	tmap1 := DeterminedMap(map[string]*Determined{"Toys": td1})
	bs, err = json.MarshalIndent(tmap1, "", "  ")
	if err != nil { t.Fatal(err) }
	if string(bs) != data0 {
	t.Errorf("%s", bs)
	t.Errorf("%s", data0)
	}

	dmap = DeterminedMap{}
	err = json.Unmarshal([]byte(data0), &dmap)
	if err != nil { t.Fatal(err) }
	ref = map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Square": &Square{}, "Toy": &Toy{}}

	adult := new(Adult)
	err = JsonUnmarshal([]byte(data1), adult, dmap, ref)
	if err != nil { t.Fatal(err) }
	_, ok := adult.Toys[0].Shape.(*Circle)
	if !ok {
		t.Errorf("%#v", adult.Toys[0])
	}
	_, ok = adult.Toys[1].Shape.(*Square)
	if !ok {
		t.Errorf("%#v", adult.Toys[1])
	}

	data1 = `{
"family":true,
"lastname":"Bizhang",
"toys":[{
	"toy_name":"roblox",
	"price":99.9,
	"geo":{
		"name": "peter shape",
		"shape": {
    		"radius":1.234
		}
	}
},{
	"toy_name":"minecraft",
	"price":199.9,
	"geo":{
		"name": "marcus shape",
		"shape": {
    		"radius":3.4
		}
	}
}]
}`
	data0 = `{
  "Toys": {
    "meta_type": 2,
    "single_name": "Toy",
    "single_field": {
      "Geo": {
        "meta_type": 1,
        "single_name": "Geo",
        "single_field": {
          "Shape": {
            "meta_type": 1,
            "single_name": "Circle"
          }
        }
      }
    }
  }
}`

	adult = new(Adult)
	err = JJUnmarshal([]byte(data1), adult, []byte(data0), ref)
	if err != nil { t.Fatal(err) }
	_, ok = adult.Toys[0].Shape.(*Circle)
	if !ok || adult.Toys[0].ToyName != "roblox" {
		t.Errorf("%#v", adult.Toys[0])
	}
	_, ok = adult.Toys[1].Shape.(*Circle)
	if !ok || adult.Toys[1].ToyName != "minecraft" {
		t.Errorf("%#v", adult.Toys[1])
	}
}
