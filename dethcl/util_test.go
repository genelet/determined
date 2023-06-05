package dethcl

import (
	"testing"
)

func TestClone(t *testing.T) {
	type Guess struct {
		A bool
		B string
	}
	type What struct {
		X string
		Y int
		Z *Guess
	}

	obj := &What{"mr x", 123, &Guess{true, "john"}}
	obj1 := clone(obj).(*What)
	if obj.X != obj1.X ||
		obj.Y != obj1.Y ||
		obj.Z.A != obj1.Z.A ||
		obj.Z.B != obj1.Z.B {
		t.Errorf("%#v => %#v\n", obj, obj1)
	}

	obj.Y = 456
	if obj1.Y == obj.Y {
		t.Errorf("%#v => %#v\n", obj, obj1)
	}
}

type some struct {
	L1 string `json:"l1" hcl:"l1,label"`
	L2 string `json:"l2" hcl:"l2,label"`
	L3 string `json:"l3" hcl:"l3,label"`
	SX int    `json:"sx" hcl:"sx"`
	SY int    `json:"sy" hcl:"sy"`
}

func TestAddLabels(t *testing.T) {
	v := &some{SX: 1, SY: 2}
	addLables(v, "a1", "b1", "c1")
	if v.L1 != "a1" || v.L2 != "b1" || v.L3 != "c1" || v.SX != 1 || v.SY != 2 {
		t.Errorf("%#v", v)
	}
}
