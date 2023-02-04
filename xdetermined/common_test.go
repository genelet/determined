package xdetermined

import (
	"testing"
)

/*
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
        Name  string            `json:"name"`
        SingleShape I           `json:"shape"`
        ListShapes []I          `json:"list_shapes"`
        HashShapes map[string]I `json:"hash_shapes"`
}
*/

func TestCommon(t *testing.T) {
	endpoint, err := NewSingleStruct(
		"Geo", map[string]interface{}{
			"TheString": "Circle",
			"TheList888":[]string{"CircleClass1", "CircleClass2"},
			"TheList":   [][2]interface{}{
				[2]interface{}{"CircleClass1"},
				[2]interface{}{"CircleClass2"}},
			"TheHash":   map[string][2]interface{}{
				"a1":[2]interface{}{"CircleClass1"},
				"b1":[2]interface{}{"CircleClass2"}},
			"TheHash888":map[string]string{
				"a1":"CircleClass1",
				"a2":"CircleClass2"},

			"Shape":     [2]interface{}{
				"Class1", map[string]interface{}{"Field1": "Circle"}},
			"ListShapes": [][2]interface{}{
				[2]interface{}{"Class2", map[string]interface{}{"Field3":"Circle"}},
				[2]interface{}{"Class3", map[string]interface{}{"Field5":"Circle"}}},
			"HashShapes": map[string][2]interface{}{
				"x1":[2]interface{}{"Class5", map[string]interface{}{"Field4":"Circle"}},
				"y1":[2]interface{}{"Class6", map[string]interface{}{"Field5":"Circle"}}},
			},
		)
	if err != nil { panic(err) }
	t.Errorf("type of endpoint: %T", endpoint)
	t.Errorf("%s", endpoint.String())
}
