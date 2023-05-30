package det

import (
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"strings"
	"testing"
)

type I interface {
	Area() float32
}

type Square struct {
	SX int `json:"sx" hcl:"sx"`
	SY int `json:"sy" hcl:"sy"`
}

func (self *Square) Area() float32 {
	return float32(self.SX * self.SY)
}

type Circle struct {
	Radius float32 `json:"radius" hcl:"radius"`
}

func (self *Circle) Area() float32 {
	return 3.14159 * self.Radius
}

type Cubic struct {
	Size int `json:"size" hcl:"size"`
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
	Name  string `json:"name" hcl:"name"`
	Shape I      `json:"shape" hcl:"shape,remain"`
}

type Geometry struct {
	Name   string       `json:"name" hcl:"name"`
	Shapes map[string]I `json:"shapes" hcl:"shapes,block"`
}

type Picture struct {
	Name     string `json:"name" hcl:"name"`
	Drawings []I    `json:"drawings" hcl:"drawings,remain"`
}

func TestSliceForMap(t *testing.T) {
	data3 := `{
	"name": "peter shapes",
	"shapes": {
		"obj5" : { "sx":5, "sy":6 },
		"obj7" : { "sx":7, "sy":8 }
	}
}`
	geometry := &Geometry{}
	spec, err := NewStruct(
		"Geometry", map[string]interface{}{
			"Shapes": []string{
				"Square"}}) // in case of key is unknown, use slice
	ref := map[string]interface{}{"Circle": new(Circle), "Square": new(Square)}
	err = JsonUnmarshal([]byte(data3), geometry, spec, ref)
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
	ref := map[string]interface{}{"Circle": c}
	spec, err := NewStruct(
		"Geo", map[string]interface{}{"Shape": "Circle"})
	err = JsonUnmarshal([]byte(data1), geo, spec, ref)
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
	ref = map[string]interface{}{"Circle": c, "Square": s}
	spec, err = NewStruct(
		"Geo", map[string]interface{}{"Shape": "Square"})
	err = JsonUnmarshal([]byte(data2), geo, spec, ref)
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
	spec, err = NewStruct(
		"Geometry", map[string]interface{}{
			"Shapes": map[string]string{
				"obj5": "Square",
				"obj7": "Square"}})
	err = JsonUnmarshal([]byte(data3), geometry, spec, ref)
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
	spec, err = NewStruct(
		"Geometry", map[string]interface{}{
			"Shapes": map[string]string{
				"obj7": "Square"}}) // in case of less items, use the first one
	err = JsonUnmarshal([]byte(data3), geometry, spec, ref)
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

	data4 := `{
	"name": "peter drawings",
	"drawings": [
		{ "sx":5, "sy":6 },
		{ "sx":7, "sy":8 }
	]
}`
	picture := &Picture{}
	spec, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"Square", "Square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = JsonUnmarshal([]byte(data4), picture, spec, ref)
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
	spec, err = NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"Square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = JsonUnmarshal([]byte(data4), picture, spec, ref)
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

	picture = &Picture{}
	err = json.Unmarshal([]byte(data4), picture)
	if err == nil || err.Error() != "json: cannot unmarshal object into Go struct field Picture.drawings of type det.I" {
		t.Fatal(err)
	}
}

type Toy struct {
	Geo     `json:"geo" hcl:"geo,block"`
	ToyName string  `json:"toy_name" hcl:"toy_name"`
	Price   float32 `json:"price" hcl:"price"`
}

func (self *Toy) ImportPrice(rate float32) float32 {
	return rate * 0.7 * self.Price
}

type Child struct {
	Toy `json:"toy" hcl:"toy,block"`
	Age int `json:"age" hcl:"age"`
}

type Child1 struct {
	Brand *Toy `json:"brand" hcl:"brand,block"`
	Age   int  `json:"age" hcl:"age"`
}

type Adult struct {
	Toys     []*Toy `json:"toys", hcl:"toys,block"`
	Family   bool   `json:"family" hcl:"family"`
	Lastname string `json:"lastname" hcl:"lastname"`
	spec     *Struct
	ref      map[string]interface{}
}

func (self *Adult) Assign(spec *Struct, ref map[string]interface{}) {
	self.spec = spec
	self.ref = ref
}
func (self *Adult) UnmarshalJSON(dat []byte) error {
	return JsonUnmarshal(dat, self, self.spec, self.ref)
}

func TestJsonToy1(t *testing.T) {
	data1 := `{
"age":5,
"brand":{
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
	spec, err := NewStruct(
		"Child", map[string]interface{}{
			"Brand": [2]interface{}{
				"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Circle"}}}}})
	ref := map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

	child := new(Child1)
	err = JsonUnmarshal([]byte(data1), child, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if child.Age != 5 || child.Brand.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", child)
	}
}

func TestJsonToy(t *testing.T) {
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
	spec, err := NewStruct(
		"Child", map[string]interface{}{
			"Toy": [2]interface{}{
				"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Circle"}}}}})
	ref := map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

	child := new(Child)
	err = JsonUnmarshal([]byte(data1), child, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	if child.Age != 5 || child.Toy.Shape.(*Circle).Radius != 1.234 {
		t.Errorf("%#v", child)
	}

	data2 := `{
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
	spec, err = NewStruct(
		"Adult", map[string]interface{}{
			"Toys": [][2]interface{}{
				{"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Circle"}}}},
				{"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Square"}}}},
			}})
	if err != nil {
		t.Fatal(err)
	}
	ref = map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Square": &Square{}, "Toy": &Toy{}}

	adult := new(Adult)
	err = JsonUnmarshal([]byte(data2), adult, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := adult.Toys[0].Shape.(*Circle)
	if !ok {
		t.Errorf("%#v", adult.Toys[0])
	}
	_, ok = adult.Toys[1].Shape.(*Square)
	if !ok {
		t.Errorf("%#v", adult.Toys[1])
	}

	second := new(Adult)
	second.Assign(spec, ref)
	err = json.Unmarshal([]byte(data2), second)
	if err != nil {
		panic(err)
	}

	_, ok = second.Toys[0].Shape.(*Circle)
	if !ok || second.Toys[0].ToyName != "roblox" {
		t.Errorf("%#v", second.Toys[0])
	}
	_, ok = second.Toys[1].Shape.(*Square)
	if !ok || second.Toys[1].ToyName != "minecraft" {
		t.Errorf("%#v", second.Toys[1])
	}
}

func TestJsonEncoding(t *testing.T) {
	spec, err := NewStruct(
		"Adult", map[string]interface{}{
			"Toys": [][2]interface{}{
				{"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Circle"}}}},
				{"Toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"Geo", map[string]interface{}{"Shape": "Square"}}}},
			}})
	if err != nil {
		t.Fatal(err)
	}
	bs, err := protojson.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ReplaceAll(string(bs), " ", "") != `{"ClassName":"Adult","fields":{"Toys":{"listStruct":{"listFields":[{"ClassName":"Toy","fields":{"Geo":{"singleStruct":{"ClassName":"Geo","fields":{"Shape":{"singleStruct":{"ClassName":"Circle"}}}}}}},{"ClassName":"Toy","fields":{"Geo":{"singleStruct":{"ClassName":"Geo","fields":{"Shape":{"singleStruct":{"ClassName":"Square"}}}}}}}]}}}}` {
		t.Errorf("%s", bs)
	}
}
