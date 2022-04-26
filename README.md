# determined

_determined_ unmarshal JSON string with interfaces determined at run-time.

[![GoDoc](https://godoc.org/github.com/genelet/determined?status.svg)](https://godoc.org/github.com/genelet/determined)

If struct has an interface field, to unmarshal its JSON string
will fail:

```bash
json: cannot unmarshal object into Go struct field XYZ of type ABC
```

Implementatin of a customized _Unmarshaller_ could be difficult or just tedious.

Alternatively, you can use this _determined_ package.

<br /><br />
## 1. Installation

To download, 

```bash
go get github.com/genelet/determined
```

<br /><br />
## 2. Usage

To use

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
    err = JJUnmarshal([]byte(data1), adult, []byte(data0), ref)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Toy 0: %#v\n", adult.Toys[0])
	fmt.Printf("Toy 1: %#v\n", adult.Toys[1])
}
```

	
<br /><br />

[Check the document](https://godoc.org/github.com/genelet/determined)
