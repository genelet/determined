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

	// clone creates a new zero-value instance of the same type
	obj := &What{"mr x", 123, &Guess{true, "john"}}
	obj1 := clone(obj).(*What)

	// Clone should be a zero-value instance, not a copy
	if obj1.X != "" || obj1.Y != 0 || obj1.Z != nil {
		t.Errorf("expected zero-value clone, got %#v", obj1)
	}

	// Verify original is unchanged
	if obj.X != "mr x" || obj.Y != 123 || obj.Z == nil {
		t.Errorf("original was modified: %#v", obj)
	}

	// Verify clone is independent - modifying original doesn't affect clone
	obj.Y = 456
	if obj1.Y != 0 {
		t.Errorf("clone should be independent, got Y=%d", obj1.Y)
	}

	// Verify clone can be populated independently
	obj1.X = "new value"
	obj1.Y = 999
	if obj.X == obj1.X || obj.Y == obj1.Y {
		t.Errorf("clone should be independent from original")
	}
}
