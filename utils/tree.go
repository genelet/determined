package utils

import (
	"sync"

	"github.com/zclconf/go-cty/cty"
)

const (
	VAR        = "var"
	ATTRIBUTES = "attributes"
	FUNCTIONS  = "functions"
)

type Tree struct {
	mu    sync.Mutex
	Name  string
	Data  map[string]cty.Value
	Up    *Tree
	Downs []*Tree
}

func NewTree(name string) *Tree {
	return &Tree{Name: name, Data: make(map[string]cty.Value)}
}

func (self *Tree) SimpleMap() map[string]interface{} {
	x := map[string]interface{}{"name": self.Name}
	if self.Up != nil {
		x["up"] = self.Up.Name
	}
	if len(self.Downs) > 0 {
		var downs []map[string]interface{}
		for _, down := range self.Downs {
			downs = append(downs, down.SimpleMap())
		}
		x["downs"] = downs
	}
	return x
}

func (self *Tree) AddNode(name string) *Tree {
	first := grep(self.Downs, name)
	if first != nil {
		return first
	}

	self.mu.Lock()
	child := NewTree(name)
	self.Downs = append(self.Downs, child)
	child.Up = self
	self.mu.Unlock()
	return child
}

func (self *Tree) AddNodes(tag string, names ...string) *Tree {
	node := self.AddNode(tag)
	for _, name := range names {
		node = node.AddNode(name)
	}
	return node
}

func grep(downs []*Tree, name string) *Tree {
	for _, down := range downs {
		if down.Name == name {
			return down
		}
	}
	return nil
}

func (self *Tree) GetNode(tag string, names ...string) *Tree {
	if tag == "" {
		return self
	}
	down := grep(self.Downs, tag)
	if down == nil {
		return nil
	}

	for _, name := range names {
		down = grep(down.Downs, name)
		if down == nil {
			return nil
		}
	}

	return down
}

func (self *Tree) DeleteNode(name string) {
	self.mu.Lock()
	for i, item := range self.Downs {
		if item.Name == name {
			if i+1 == len(self.Downs) {
				self.Downs = self.Downs[:i]
			} else {
				self.Downs = append(self.Downs[:i], self.Downs[i+1:]...)
			}
			self.mu.Unlock()
			return
		}
	}
	self.mu.Unlock()
}

func (self *Tree) AddItem(k string, v cty.Value) {
	self.mu.Lock()
	self.Data[k] = v
	self.mu.Unlock()
}

func (self *Tree) DeleteItem(k string) {
	self.mu.Lock()
	delete(self.Data, k)
	self.mu.Unlock()
}

func (self *Tree) FindNode(names []string) *Tree {
	if names == nil || len(names) == 0 {
		return nil
	}

	var down *Tree

	for _, item := range self.Downs {
		if item.Name == names[0] {
			if len(names) == 1 {
				return item
			}
			return item.FindNode(names[1:])
		} else {
			down = item.FindNode(names)
			if down != nil {
				return down
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
