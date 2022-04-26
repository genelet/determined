# determined

_Determined_ unmarshal JSON string with interfaces determined at run-time.

[![GoDoc](https://godoc.org/github.com/genelet/determined?status.svg)](https://godoc.org/github.com/genelet/determined)

If a struct contains interface field, unmarshaller using _encoding/json_ will fail with the error:

```bash
json: cannot unmarshal object into Go struct field XYZ of type ABC
```

This package helps you to implement customized _Unmarshaller_ easily.

Alternatively, you can use it directly.

## 1. Installation

To download, 

```bash
go get github.com/genelet/determined
```

## 2. Usage

In the following example, _Adult_ contains multiple-level objects including an interface. We create a customized marshaller using _determined_
and then _json.Unmarshal_ works again.

```go
package main

import (
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

type Adult struct {
    Toys []*Toy `json:"toys"`
    Family bool `json:"family"`
    Lastname string `json:"lastname"`
    objectMap []byte
    ref map[string]interface{}
}
func (self *Adult) Assign(objectMap []byte, ref map[string]interface{}) {
    self.objectMap = objectMap
    self.ref = ref
}
func (self *Adult) UnmarshalJSON(dat []byte) error {
    return JJUnmarshal(dat, self, self.objectMap, self.ref)
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

// carefully build this DeterminedMap:
   data0 = `{
"Toys": {"meta_type": 4, "slice_name": ["Toy", "Toy"], "slice_field": [{
    "Geo": {"meta_type": 1, "single_name": "Geo", "single_field": {
        "Shape": {"meta_type": 1, "single_name": "Circle"}}}
    },{
    "Geo": {"meta_type": 1, "single_name": "Geo", "single_field": {
        "Shape": {"meta_type": 1, "single_name": "Square"}}}
    }]
}}`
    ref = map[string]interface{}{"Geo": &Geo{}, "Toy": &Toy{}, "Circle": &Circle{}, "Square": &Square{}}
    
    adult := new(Adult)
    adult.Assign([]byte(data0), ref)
    err = json.Unmarshal([]byte(data1), adult)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Toy 0: %#v\n", adult.Toys[0])
    fmt.Printf("Toy 1: %#v\n", adult.Toys[1])
}
```

The output will looks like:
```bash
```
	
[Check the document](https://godoc.org/github.com/genelet/determined)
