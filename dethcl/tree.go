package dethcl

import (
	"github.com/zclconf/go-cty/cty"
)

const (
	VAR        = "var"
	ATTRIBUTES = "attributes"
	FUNCTIONS  = "functions"
)

type Tree struct {
	Name  string
	Data  map[string]cty.Value
	Up    *Tree
	Downs []*Tree
}

func NewTree(name string) *Tree {
	return &Tree{Name: name, Data: make(map[string]cty.Value)}
}

func (self *Tree) AddNode(name string) *Tree {
	child := NewTree(name)
	self.Downs = append(self.Downs, child)
	child.Up = self
	return child
}

func (self *Tree) DeleteNode(name string) {
	for i, item := range self.Downs {
		if item.Name == name {
			if i+1 == len(self.Downs) {
				self.Downs = self.Downs[:i]
			} else {
				self.Downs = append(self.Downs[:i], self.Downs[i+1:]...)
			}
			return
		}
	}
}

func (self *Tree) AddItem(k string, v cty.Value) {
	self.Data[k] = v
}

func (self *Tree) DeleteItem(k string) {
	delete(self.Data, k)
}

func (self *Tree) FindNode(names []string) *Tree {
	if names == nil || len(names) == 0 {
		return self
	}
	for _, item := range names {
		for _, down := range self.Downs {
			if down.Name == item {
				if len(names) == 1 {
					return down
				}
				return down.FindNode(names[1:])
			}
		}
	}
	return nil
}

// this type of Variable name is not implemented in ScopeTraversalExpr
// but could be used in function
// note: hcl.Traversal.TraverseAbs uses only Variables in Eval
func (self *Tree) Variables() map[string]cty.Value {
	output := make(map[string]cty.Value)
	for k, v := range self.Data {
		output[self.Name+"."+k] = v
		if self.Name == VAR {
			output[k] = v
		}
	}
	for _, down := range self.Downs {
		for k, v := range down.Variables() {
			output[self.Name+"."+k] = v
		}
	}
	if len(output) == 1 {
		return nil
	}
	return output
}
