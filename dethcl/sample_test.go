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

type team struct {
	TeamName string `json:"team_name" hcl:"team_name,label"`
	SX       int    `json:"sx" hcl:"sx"`
	SY       int    `json:"sy" hcl:"sy"`
}

func (self *team) Area() float32 {
	return float32(self.SX * self.SY)
}

type moresquare struct {
	Morename1 string `json:"morename1" hcl:"morename1,label"`
	Morename2 string `json:"morename2" hcl:"morename2,label"`
	SX        int    `json:"sx" hcl:"sx"`
	SY        int    `json:"sy" hcl:"sy"`
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
	Radius float32  `json:"radius" hcl:"radius"`
	Arr1   []string `json:"arr1" hcl:"arr1,attr"`
	Arr2   []int32  `json:"arr2" hcl:"arr2,attr"`
	Arr3   []bool   `json:"arr3" hcl:"arr3,attr"`
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
	Name     string  `json:"name" hcl:"name"`
	Drawings []inter `json:"drawings" hcl:"drawings,block"`
}

type config struct {
	Name     string  `json:"name" hcl:"name"`
	Drawings []inter `json:"drawings" hcl:"drawings,block"`
}

type Painting struct {
	Name     string  `json:"name" hcl:"name"`
	Drawings []inter `json:"drawings" hcl:"drawings,block"`
}

func (self *Painting) UnmarshalHCL(dat []byte, labels ...string) error {
	spec, err := NewStruct(
		"Painting", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		return err
	}
	g := &geo{}
	s := &square{}
	c := &circle{}
	ref := map[string]interface{}{"geo": g, "circle": c, "square": s}

	return UnmarshalSpec(dat, self, spec, ref)
}

type gallery struct {
	City   string               `json:"city" hcl:"city"`
	Makers map[string]*Painting `json:"makers" hcl:"makers,block"`
}

type X7 struct {
	Many int    `json:"many" hcl:"many,optional"`
	Why  string `json:"why" hcl:"why,optional"`
}

type X8 struct {
	Number int    `json:"number" hcl:"number,optional"`
	What   string `json:"what" hcl:"what,optional"`
}

type frame0 struct {
	FrameName   string            `json:"fname" hcl:"fname,label"`
	Description string            `json:"description" hcl:"description,optional"`
	Y7          *X7               `json:"y6" hcl:"y7,block"`
	Y10         []*X7             `json:"y10" hcl:"y10,block"`
	Y11         map[string]*X7    `json:"y11" hcl:"y11,optional"`
	Y12         map[string]X7     `json:"y12" hcl:"y12,optional"`
	Y13         map[string]string `json:"y13" hcl:"y13,optional"`
	Y14         map[string]int    `json:"y14" hcl:"y14,optional"`
	Y15         map[string]bool   `json:"y15" hcl:"y15,optional"`
}
type frame struct {
	FrameName   string          `json:"fname" hcl:"fname,label"`
	Description string          `json:"description" hcl:"description,optional"`
	X1          *geo            `json:"x1" hcl:"x1,block"`
	X2          geo             `json:"x2" hcl:"x2,block"`
	X3          []*geo          `json:"x3" hcl:"x3,block"`
	X4          []geo           `json:"x4" hcl:"x4,block"`
	X5          map[string]*geo `json:"x5" hcl:"x5,block"`
	X6          map[string]geo  `json:"x6" hcl:"x6,block"`
	Y7          *X7             `json:"y6" hcl:"y7,block"`
	X8
	Y10 []*X7          `json:"y10" hcl:"y10,block"`
	Y11 map[string]*X7 `json:"y11" hcl:"y11,optional"`
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
	Age   int  `json:"age" hcl:"age"`
}
