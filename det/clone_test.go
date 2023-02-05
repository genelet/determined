package det

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
