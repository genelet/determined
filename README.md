# determined

_Determined_ unmarshals JSON string to _go struct_ containing interfaces determined at run-time.

[![GoDoc](https://godoc.org/github.com/genelet/determined?status.svg)](https://godoc.org/github.com/genelet/determined)

If object contains interface fields, unmarshaling by _encoding/json_ will result in the error:

```bash
json: cannot unmarshal object into Go struct field XYZ of type ABC
```

The solution is to build customized _Unmarshaler_ for the object. This package helps to implement [Unmarshaler](https://pkg.go.dev/encoding/json@go1.18.1#Unmarshaler) easily.

We may also use functions [JJUnmarshal](https://pkg.go.dev/github.com/genelet/determined#JJUnmarshal) or [JsonUnmarshal](https://pkg.go.dev/github.com/genelet/determined#JsonUnmarshal) directly.

## 1. Installation

To download, 

```bash
go get github.com/genelet/determined
```

## 2. Usage

In the following example, _Child_ contains multiple-level objects including an interface in _Geo_. To unmarshal a JSON to _Child_ by _json.Unmarshal_ does not work.

We need to build _UnmarshalJSON_ for _Child_.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/genelet/determined"
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

type Geo struct {
    Name  string `json:"name"`
    Shape I      `json:"shape"`
}

type Toy struct {
    Geo     `json:"geo"`
    ToyName string  `json:"toy_name"`
    Price   float32 `json:"price"`
}

type Child struct {
    Toys []*Toy `json:"toys"`
    Family bool `json:"family"`
    Lastname string `json:"lastname"`
    endpoint []byte
    ref map[string]interface{}
}
func (self *Child) Assign(endpoint []byte, ref map[string]interface{}) {
    self.endpoint = endpoint
    self.ref = ref
}
func (self *Child) UnmarshalJSON(dat []byte) error {
    return xxx.JJUnmarshal(dat, self, self.endpoint, self.ref)
}

func main() {
    data1 := `{
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

// build this Determined:
    data0 = `{"meta_type": 1, "single_name": "Child", "single_field": {
"Toys": {"meta_type": 4, "slice_name": ["Toy", "Toy"], "slice_field": [{
    "Geo": {"meta_type": 1, "single_name": "Geo", "single_field": {
        "Shape": {"meta_type": 1, "single_name": "Circle"}}}
    },{
    "Geo": {"meta_type": 1, "single_name": "Geo", "single_field": {
        "Shape": {"meta_type": 1, "single_name": "Square"}}}
    }]
}}}`
    ref := map[string]interface{}{"Geo": &Geo{}, "Circle": &Circle{}, "Square": &Square{}, "Toy": &Toy{}}
    child := new(Child)
    child.Assign([]byte(data0), ref)
    
    err := json.Unmarshal([]byte(data1), child)
    if err != nil { panic(err) }

    fmt.Printf("Toy 0: %#v\n", child.Toys[0])
    fmt.Printf("Toy 1: %#v\n", child.Toys[1])
}
```

The output will look like:
```bash
Toy 0: &main.Toy{Geo:main.Geo{Name:"peter shape", Shape:(*main.Circle)(0xc0000169f0)}, ToyName:"roblox", Price:99.9}
Toy 1: &main.Toy{Geo:main.Geo{Name:"marcus shape", Shape:(*main.Square)(0xc000016b20)}, ToyName:"minecraft", Price:199.9}
```

You may also build _child's Determined_ manually by:

```go
    sd1   := &Determined{MetaType: METASingle, SingleName: "Circle"}
    smap1 := DeterminedMap(map[string]*Determined{"Shape": sd1})
    gd1   := &Determined{MetaType: METASingle, SingleName: "Geo", SingleField: smap1}
    gmap1 := DeterminedMap(map[string]*Determined{"Geo": gd1})

    sd2   := &Determined{MetaType: METASingle, SingleName: "Square"}
    smap2 := DeterminedMap(map[string]*Determined{"Shape": sd2})
    gd2   := &Determined{MetaType: METASingle, SingleName: "Geo", SingleField: smap2}
    gmap2 := DeterminedMap(map[string]*Determined{"Geo": gd2})

    td1   := &Determined{MetaType: METASlice, SliceName: []string{"Toy", "Toy"}, SliceField: []DeterminedMap{gmap1, gmap2}}
    theMap := DeterminedMap(map[string]*Determined{"Toys": td1})

    theDetermined := &Determined{MetaType: METASingle, SingleName: "Child", SingleField: theMap}
```
	
[Check the document](https://godoc.org/github.com/genelet/determined)
