package dethcl

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

type moresquare struct {
	Morename1 string `json:"morename1", hcl:"morename1,label"`
	Morename2 string `json:"morename2", hcl:"morename2,label"`
	SX int `json:"sx" hcl:"sx"`
	SY int `json:"sy" hcl:"sy"`
}

func (self *moresquare) Area() float32 {
	return float32(self.SX * self.SY)
}

type circle struct {
	Radius float32 `json:"radius" hcl:"radius"`
}

func (self *circle) Area() float32 {
	return 3.14159 * self.Radius
}

type circlemore struct {
	Radius float32 `json:"radius" hcl:"radius"`
	Arr1 []string `json:"arr1" hcl:"arr1,attr"`
	Arr2 []int32 `json:"arr2" hcl:"arr2,attr"`
	Arr3 []bool `json:"arr3" hcl:"arr3,attr"`
}

func (self *circlemore) Area() float32 {
	return 3.14159 * self.Radius
}

type cubic struct {
	Size int `json:"size" hcl:"size"`
}

func (self *cubic) Area() float32 {
	return 6.0 * float32(self.Size*self.Size)
}

type geo struct {
	Name  string `json:"name" hcl:"name"`
	Shape inter  `json:"shape" hcl:"shape,block"`
}

type geometry struct {
	Name   string           `json:"name" hcl:"name"`
	Shapes map[string]inter `json:"shapes" hcl:"shapes,block"`
}

type picture struct {
	Name     string   `json:"name" hcl:"name"`
	Drawings []inter  `json:"drawings" hcl:"drawings,block"`
}

type toy struct {
	Geo     geo     `json:"geo" hcl:"geo,block"`
	ToyName string  `json:"toy_name" hcl:"toy_name"`
	Price   float32 `json:"price" hcl:"price"`
}

func (self *toy) ImportPrice(rate float32) float32 {
	return rate * 0.7 * self.Price
}

type child struct {
	Brand *toy `json:"brand" hcl:"brand,block"`
	Age int `json:"age" hcl:"age"`
}

