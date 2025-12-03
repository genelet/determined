package det

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/genelet/schema"
	"google.golang.org/protobuf/encoding/protojson"
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
	spec, err := schema.NewStruct(
		"Geometry", map[string]any{
			"Shapes": []string{
				"Square"}}) // in case of key is unknown, use slice
	if err != nil {
		t.Fatal(err)
	}
	ref := map[string]any{"Circle": new(Circle), "Square": new(Square)}
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
	ref := map[string]any{"Circle": c}
	spec, err := schema.NewStruct(
		"Geo", map[string]any{"Shape": "Circle"})
	if err != nil {
		t.Fatal(err)
	}
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
	ref = map[string]any{"Circle": c, "Square": s}
	spec, err = schema.NewStruct(
		"Geo", map[string]any{"Shape": "Square"})
	if err != nil {
		t.Fatal(err)
	}
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
	spec, err = schema.NewStruct(
		"Geometry", map[string]any{
			"Shapes": map[string]string{
				"obj5": "Square",
				"obj7": "Square"}})
	if err != nil {
		t.Fatal(err)
	}
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
	spec, err = schema.NewStruct(
		"Geometry", map[string]any{
			"Shapes": map[string]string{
				"obj7": "Square"}}) // in case of less items, use the first one
	if err != nil {
		t.Fatal(err)
	}
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
	spec, err = schema.NewStruct(
		"Picture", map[string]any{
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
	spec, err = schema.NewStruct(
		"Picture", map[string]any{
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
	Toys     []*Toy `json:"toys" hcl:"toys,block"`
	Family   bool   `json:"family" hcl:"family"`
	Lastname string `json:"lastname" hcl:"lastname"`
	spec     *schema.Struct
	ref      map[string]any
}

func (self *Adult) Assign(spec *schema.Struct, ref map[string]any) {
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
	spec, err := schema.NewStruct(
		"Child", map[string]any{
			"Brand": [2]any{
				"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Circle"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	ref := map[string]any{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

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
	spec, err := schema.NewStruct(
		"Child", map[string]any{
			"Toy": [2]any{
				"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Circle"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	ref := map[string]any{"Geo": &Geo{}, "Circle": &Circle{}, "Toy": &Toy{}}

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
	spec, err = schema.NewStruct(
		"Adult", map[string]any{
			"Toys": [][2]any{
				{"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Circle"}}}},
				{"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Square"}}}},
			}})
	if err != nil {
		t.Fatal(err)
	}
	ref = map[string]any{"Geo": &Geo{}, "Circle": &Circle{}, "Square": &Square{}, "Toy": &Toy{}}

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
		t.Fatal(err)
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
	spec, err := schema.NewStruct(
		"Adult", map[string]any{
			"Toys": [][2]any{
				{"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Circle"}}}},
				{"Toy", map[string]any{
					"Geo": [2]any{
						"Geo", map[string]any{"Shape": "Square"}}}},
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

type NoTagStruct struct {
	NoTagInt int
	Name     string `json:"name"`
}

func TestNoTagFieldPanic(t *testing.T) {
	s := &NoTagStruct{}
	data := []byte(`{"name": "test"}`)
	spec, err := schema.NewStruct("NoTagStruct")
	if err != nil {
		t.Fatal(err)
	}

	// This call triggers loopFields. If the bug exists, it will panic because NoTagInt has no tag and is not a struct.
	err = JsonUnmarshal(data, s, spec, nil)
	if err != nil {
		t.Fatal(err)
	}

	if s.Name != "test" {
		t.Errorf("expected name to be 'test', got %s", s.Name)
	}
}
