package utils

import (
	"testing"
)

func TestTree(t *testing.T) {
	top := NewTree("var")
	first := top.AddNode("firstLevel")
	first.AddNode("second1")
	second2 := first.AddNode("second2")

	tree := top.FindNode([]string{"firstLevel", "second1"})
	if tree == nil {
		t.Fatal("not found")
	}
	if tree.Name != "second1" {
		t.Errorf("%#v", tree)
	}
	first.AddNode("second3")
	first.DeleteNode("second2")
	if first.Downs[0].Name != "second1" || first.Downs[1].Name != "second3" {
		t.Errorf("%#v", first)
		t.Errorf("%#v", second2)
	}

	first.AddNode("second4")
	first.DeleteNode("second4")
	if first.Downs[0].Name != "second1" ||
		first.Downs[1].Name != "second3" ||
		len(first.Downs) != 2 {
		t.Errorf("%#v", first)
	}
}
