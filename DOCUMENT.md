# Decoding of Dynamic JSON Data
<br>

## 1. Introduction

The core Golang package  [_encoding/json_](https://pkg.go.dev/encoding/json)  is an exceptional library for managing JSON data. However, to decode the interface type, it necessitates writing a customized  _Unmarshaler_  for the target object. While this isn’t typically a challenging task, it often results in repetitive code for different types of objects and packages.

Therefore,  _determined_  was created to streamline the coding process and enhance productivity.

## 2. Protobuf

Following the idea in  [_reflect_](https://pkg.go.dev/reflect), we use  [protobuf](https://protobuf.dev/)  to implement  _Unmarshaler._

The following proto interprets an interface type in dynamic JSON data:
```bash
    syntax = "proto3";  
      
    package det;  
      
    option go_package = "./det";  
      
    message Struct {  
     string class_name = 1;  
     map<string, Value> fields = 2;  
    }  
      
    message Value {  
      // The kind of value.  
      oneof kind {  
        Struct single_struct   = 1;  
        ListStruct list_struct = 2;  
        MapStruct map_struct   = 3;  
      }  
    }  
      
    message ListStruct {  
      repeated Struct list_fields = 1;  
    }  
      
    message MapStruct {  
      map<string, Struct> map_fields = 1;  
    }
```
where c_lass_name  is the  _go struct_  type name at run-time.

The CLI,  _protoc_  will generate the following Golang code:
```go
    type Struct struct {  
     ClassName string            `protobuf:"bytes,1,opt,name=ClassName,proto3" json:"ClassName,omitempty"`  
     Fields    map[string]*Value `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`  
    }  
      
    type Value struct {  
     // The kind of value.  
     //  
     // Types that are assignable to Kind:  
     // *Value_SingleStruct  
     // *Value_ListStruct  
     // *Value_MapStruct  
     Kind isValue_Kind `protobuf_oneof:"kind"`  
    }  
    
    ....
```

## 3. _NewStruct_

There’s no need for users to interact directly with the aforementioned proto-generated package. Instead, the required step involves creating a Golang map to interpret the interface. This map is then passed to the  _NewStruct_  function to generate a new  _Struct:_
```go
    // NewStruct constructs a Struct from a generic Go map.  
    // The map keys of v must be valid UTF-8.  
    // The map values of v are converted using NewValue.  
    func NewStruct(name string, v ...map[string]interface{}) (*Struct, error)  
    //   
    // NewValue conversion:  
    // ╔═══════════════════════════╤══════════════════════════════╗  
    // ║ Go type                   │ Conversion                   ║  
    // ╠═══════════════════════════╪══════════════════════════════╣  
    // ║ string                    │ ending SingleStruct value    ║  
    // ║ []string                  │ ending ListStruct value      ║  
    // ║ map[string]string         │ ending MapStruct value       ║  
    // ║                           │                              ║  
    // ║ [2]interface{}            │ SingleStruct value           ║  
    // ║ [][2]interface{}          │ ListStruct value             ║  
    // ║ map[string][2]interface{} │ MapStruct value              ║  
    // ║                           │                              ║  
    // ║ *Struct                   │ SingleStruct                 ║  
    // ║ []*Struct                 │ ListStruct                   ║  
    // ║ map[string]*Struct        │ MapStruct                    ║  
    // ╚═══════════════════════════╧══════════════════════════════╝
```
Fields of primitive data types, or with defined  _go struct,_  should be ignored, since they will be decoded automatically by  _encoding/json_.

Here are examples of  _NewStruct_.

**Single Interface:**

Here  _geo_  contains interface field  `Shape`.
```go
    type geo struct {  
        Name  string `json:"name" hcl:"name"`  
        Shape inter  `json:"shape" hcl:"shape,block"`  
    }  
      
    type inter interface {  
        Area() float32  
    }  
      
    type square struct {  
        SX int `json:"sx" hcl:"sx"`  
        SY int `json:"sy" hcl:"sy"`  
    }  
      
    func (self *square) Area() float32 {  
        return float32(self.SX * self.SY)  
    }  
      
    type circle struct {  
        Radius float32 `json:"radius" hcl:"radius"`  
    }  
      
    func (self *circle) Area() float32 {  
        return 3.14159 * self.Radius  
    }
```
Assume a serialized JSON of  _geo_  is received, where field  `Shape` is known to be  _circle_  at run-time. To decode  `Shape`  of type _circle_, build the following  _spec_:
```go
    spec, err := NewStruct(  
      "geo", map[string]interface{}{"Shape": "circle"})
```
If  `Shape`  is type  _square,_  build this  _spec_:
```go
    spec, err = NewStruct(  
      "geo", map[string]interface{}{"Shape": "square"})
```
**List of interface:**

Here  _picture_  contains  `Drawings`, which is a slice of interface. The  _spec_  for serialized slice of  _square_,  _circle_  or combination of  _square_  and  _circle_, are
```go
    type picture struct {  
    Name string  `json:"name" hcl:"name"`  
    Drawings []inter `json:"drawings" hcl:"drawings,block"`  
    }  
```
 
incoming data is slice of square, size 2 :
```go
    spec, err := NewStruct(  
    "Picture", map[string]interface{}{  
    "Drawings": []string{"square", "square"}})  
```
  
the first element is square and the second circle  :
```go
    spec, err := NewStruct(  
    "Picture", map[string]interface{}{  
    "Drawings": []string{"square", "circle"}})  
```
  
if all elements of interface slice is type square :
```go
    spec, err := NewStruct(  
    "Picture", map[string]interface{}{  
    "Drawings": []string{"square"}})
```
 

Note that if all elements are of the same type  _square_  in  `Drawing`, just pass 1-element array  _[]string{“square”}._

**Map of interface:**

Here  `Shapes`  is a map of interface:
```go
    type geometry struct {  
        Name   string           `json:"name" hcl:"name"`  
        Shapes map[string]inter `json:"shapes" hcl:"shapes,block"`  
    }  
      
    spec, err := NewStruct(  
      "geometry", map[string]interface{}{  
        "Shapes": map[string]string{"k1":"square", "k2":"square"}})  
```
  
if all values of interface map is type square  :
```go
    spec, err := NewStruct(  
      "Picture", map[string]interface{}{  
        "Shapes": []string{"square"}})
```
**Nested:**

In  _toy_,  `Geo`  is of type  _geo_  which contains interface  `Shape`:
```go
    type toy struct {  
        Geo     geo     `json:"geo" hcl:"geo,block"`  
        ToyName string  `json:"toy_name" hcl:"toy_name"`  
        Price   float32 `json:"price" hcl:"price"`  
    }  
      
    spec, err = NewStruct(  
      "toy", map[string]interface{}{  
        "Geo": [2]interface{}{  
          "geo", map[string]interface{}{"Shape": "square"}}})
```
**Nested of nested:**

Here  _child_  has field  `Brand`  which is a map of nested  _toy_:
```go
    type child struct {  
        Brand map[string]*toy `json:"brand" hcl:"brand,block"`  
        Age   int  `json:"age" hcl:"age"`  
    }  
      
    spec, err = NewStruct(  
      "child", map[string]interface{}{  
        "Brand": [][2]interface{}{  
          "k1":[2]interface{}{"toy", map[string]interface{}{  
            "Geo": [2]interface{}{  
              "geo", map[string]interface{}{"Shape": "circle"}}}},  
          "k2":[2]interface{}{"toy", map[string]interface{}{  
            "Geo": [2]interface{}{  
              "geo", map[string]interface{}{"Shape": "square"}}}},  
        },  
      },  
    )
```
## 4. Use JsonUnmarshal to Decode JSON

To decode JSON to object containing interface types, use  _JsonUnmarshal_:

```go
// JsonUnmarshal unmarshals JSON data with interfaces determined by spec.
//
//   - dat: JSON data
//   - current: object as pointer
//   - spec: *Struct
//   - ref: struct map, with key being string name and value reference to struct
func JsonUnmarshal(dat []byte, current interface{}, spec *Struct, ref map[string]interface{}) error
```    
The following program decodes JSON  _data1_  into object  _child_:
 ```go   
package main

import (
    "fmt"

    "github.com/genelet/determined/det"
)

type geo struct {
    Name  string `json:"name" hcl:"name"`
    Shape inter  `json:"shape" hcl:"shape,block"`
}

type inter interface {
    Area() float32
}

type square struct {
    SX int `json:"sx" hcl:"sx"`
    SY int `json:"sy" hcl:"sy"`
}

func (self *square) Area() float32 {
    return float32(self.SX * self.SY)
}

type circle struct {
    Radius float32 `json:"radius" hcl:"radius"`
}

func (self *circle) Area() float32 {
    return 3.14159 * self.Radius
}

type toy struct {
    Geo     geo     `json:"geo" hcl:"geo,block"`
    ToyName string  `json:"toy_name" hcl:"toy_name"`
    Price   float32 `json:"price" hcl:"price"`
}

type child struct {
    Brand map[string]*toy `json:"brand" hcl:"brand,block"`
    Age   int  `json:"age" hcl:"age"`
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
        "name" : "quare shape",
        "shape" : { "sx" : 5, "sy" : 6 }
    }
}
}
}`

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
    ref := map[string]interface{}{"toy": &toy{}, "geo": &geo{}, "circle": &circle{}, "square": &square{}}

    c := new(child)
    err = det.JsonUnmarshal([]byte(data1), c, spec, ref)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%v\n", c.Age)
    fmt.Printf("%#v\n", c.Brand["abc1"])
    fmt.Printf("%#v\n", c.Brand["abc1"].Geo.Shape)
    fmt.Printf("%#v\n", c.Brand["def2"])
    fmt.Printf("%#v\n", c.Brand["def2"].Geo.Shape)
}
```
the program outputs:
```bash
    &main.toy{Geo:main.geo{Name:"medium shape", Shape:(*main.circle)(0xc0000b6468)}, ToyName:"roblox", Price:99.9}  
    &main.circle{Radius:1.234}  
    &main.toy{Geo:main.geo{Name:"square shape", Shape:(*main.square)(0xc0000b6350)}, ToyName:"minecraft", Price:9.9}  
    &main.square{SX:5, SY:6}
```

## 5. Customized Unmarshaler of encoding/json

If  _UnmarshalJSON_  is implemented on  _go struct_, it is said to have a customized  _unmarshaler_  and so Golang core package  _encoding/json_ will automatically decode it.

With  _JsonUnmarshal,_ we can easily write a customized  _unmarshaler_  for  _child_:
```go
    type child struct {  
        Brand map[string]*toy `json:"brand" hcl:"brand,block"`  
        Age   int  `json:"age" hcl:"age"`  
        spec  *det.Struct  
        ref map[string]interface{}  
    }  
    func (self *child) Assign(spec *det.Struct, ref map[string]interface{}) {  
        self.spec = spec  
        self.ref = ref  
    }  
    func (self *child) UnmarshalJSON(dat []byte) error {  
        return det.JsonUnmarshal(dat, self, self.spec, self.ref)  
    }
```
Now the sample code below can use  _encoding/json_  to decode:
```go
    package main  
      
    import (  
        "encoding/json"  
        "fmt"  
      
        "github.com/genelet/determined/det"  
    )  
      
    type geo struct {  
        Name  string `json:"name" hcl:"name"`  
        Shape inter  `json:"shape" hcl:"shape,block"`  
    }  
      
    type inter interface {  
        Area() float32  
    }  
      
    type square struct {  
        SX int `json:"sx" hcl:"sx"`  
        SY int `json:"sy" hcl:"sy"`  
    }  
      
    func (self *square) Area() float32 {  
        return float32(self.SX * self.SY)  
    }  
      
    type circle struct {  
        Radius float32 `json:"radius" hcl:"radius"`  
    }  
      
    func (self *circle) Area() float32 {  
        return 3.14159 * self.Radius  
    }  
      
    type toy struct {  
        Geo     geo     `json:"geo" hcl:"geo,block"`  
        ToyName string  `json:"toy_name" hcl:"toy_name"`  
        Price   float32 `json:"price" hcl:"price"`  
    }  
      
    type child struct {  
        Brand map[string]*toy `json:"brand" hcl:"brand,block"`  
        Age   int  `json:"age" hcl:"age"`  
        spec  *det.Struct  
        ref map[string]interface{}  
    }  
      
    func (self *child) Assign(spec *det.Struct, ref map[string]interface{}) {  
        self.spec = spec  
        self.ref = ref  
    }  
      
    func (self *child) UnmarshalJSON(dat []byte) error {  
        return det.JsonUnmarshal(dat, self, self.spec, self.ref)  
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
            "name" : "quare shape",  
            "shape" : { "sx" : 5, "sy" : 6 }  
        }  
    }  
    }  
    }`  
      
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
        ref := map[string]interface{}{"toy": &toy{}, "geo": &geo{}, "circle": &circle{}, "square": &square{}}  
      
        c := new(child)  
        c.Assign(spec, ref)  
        err = json.Unmarshal([]byte(data1), c)  
        if err != nil {  
            panic(err)  
        }  
        fmt.Printf("%v\n", c.Age)  
        fmt.Printf("%#v\n", c.Brand["abc1"])  
        fmt.Printf("%#v\n", c.Brand["abc1"].Geo.Shape)  
        fmt.Printf("%#v\n", c.Brand["def2"])  
        fmt.Printf("%#v\n", c.Brand["def2"].Geo.Shape)  
    }
```
The advantage of using a customized  _unmarshaler_  is that any Go struct, which encapsulates a child, can directly use  _encoding/json_  without worrying about interface fields in the child.

<br>

<br>
