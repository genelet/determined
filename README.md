# determined

_Determined_ marshals and unmarshals JSON data to _go struct_ containing interfaces determined at run-time.

[![GoDoc](https://godoc.org/github.com/genelet/determined?status.svg)](https://godoc.org/github.com/genelet/determined)

> **Note**: The HCL parsing feature has been extracted and moved to a separate Github package [github.com/genelet/horizon](https://github.com/genelet/horizon). The old HCL code in this package is frozen for compatibility reasons.

## Installation

```bash
go get github.com/genelet/determined
```

## Introduction

The core Golang package `encoding/json` is an exceptional library for managing JSON data. However, to decode the interface type, it necessitates writing a customized `Unmarshaler` for the target object. While this isn’t typically a challenging task, it often results in repetitive code for different types of objects and packages.

Therefore, `determined` was created to streamline the coding process and enhance productivity.

## Usage

To decode JSON to an object containing interface types, use `det.JsonUnmarshal`.

```go
// JsonUnmarshal unmarshals JSON data with interfaces determined by spec.
//
//   - dat: JSON data
//   - current: object as pointer
//   - spec: *Struct
//   - ref: struct map, with key being string name and value reference to struct
func JsonUnmarshal(dat []byte, current interface{}, spec *Struct, ref map[string]interface{}) error
```

You first need to define the structure of your interfaces using `det.NewStruct`.

### Example

Here is an example of how to use `determined` to decode JSON data where fields are interfaces.

```go
package main

import (
    "fmt"
    "github.com/genelet/determined/det"
)

type geo struct {
    Name  string `json:"name"`
    Shape inter  `json:"shape"`
}

type inter interface {
    Area() float32
}

type square struct {
    SX int `json:"sx"`
    SY int `json:"sy"`
}

func (self *square) Area() float32 {
    return float32(self.SX * self.SY)
}

type circle struct {
    Radius float32 `json:"radius"`
}

func (self *circle) Area() float32 {
    return 3.14159 * self.Radius
}

type toy struct {
    Geo     geo     `json:"geo"`
    ToyName string  `json:"toy_name"`
    Price   float32 `json:"price"`
}

type child struct {
    Brand map[string]*toy `json:"brand"`
    Age   int  `json:"age"`
}

func main() {
    data1 := `{
        "age" : 5,
        "brand" : {
            "abc1" : {
                "toy_name" : "roblox",
                "price" : 99.9,
                "geo" : {
                    "name" : "medium shape",
                    "shape" : { "radius" : 1.234 }
                }
            },
            "def2" : {
                "toy_name" : "minecraft",
                "price" : 9.9,
                "geo" : {
                    "name" : "square shape",
                    "shape" : { "sx" : 5, "sy" : 6 }
                }
            }
        }
    }`

    // Define the structure of the interfaces for the specific data
    spec, err := det.NewStruct(
        "child", map[string]interface{}{
            "Brand": map[string][2]interface{}{
                "abc1":[2]interface{}{"toy", map[string]interface{}{
                    "Geo": [2]interface{}{
                        "geo", map[string]interface{}{"Shape": "circle"}}}},
                "def2":[2]interface{}{"toy", map[string]interface{}{
                    "Geo": [2]interface{}{
                        "geo", map[string]interface{}{"Shape": "square"}}}},
            },
        },
    )
    if err != nil {
        panic(err)
    }

    // Map string names to actual struct pointers
    ref := map[string]interface{}{
        "toy": &toy{}, 
        "geo": &geo{}, 
        "circle": &circle{}, 
        "square": &square{},
    }

    c := new(child)
    err = det.JsonUnmarshal([]byte(data1), c, spec, ref)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Age: %v\n", c.Age)
    fmt.Printf("Brand abc1: %#v\n", c.Brand["abc1"])
    fmt.Printf("Brand abc1 Shape: %#v\n", c.Brand["abc1"].Geo.Shape)
    fmt.Printf("Brand def2: %#v\n", c.Brand["def2"])
    fmt.Printf("Brand def2 Shape: %#v\n", c.Brand["def2"].Geo.Shape)
}
```

For more details on how to construct the `spec` using `NewStruct`, please refer to [DOCUMENT.md](DOCUMENT.md).
